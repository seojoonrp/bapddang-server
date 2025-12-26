// api/services/user_service.go

package services

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/seojoonrp/bapddang-server/api/repositories"
	"github.com/seojoonrp/bapddang-server/config"
	"github.com/seojoonrp/bapddang-server/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/api/idtoken"
)

type AppleKey struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	N string `json:"n"`
	E string `json:"e"`
}

type AppleKeys struct {
	Keys []AppleKey `json:"keys"`
}

type UserService interface {
	SignUpWithEmail(input models.SignUpInput) error
	LoginWithEmail(input models.LoginInput) (string, *models.User, error)

	LoginWithGoogle(idToken string) (string, *models.User, bool, error)
	LoginWithKakao(accessToken string) (string, *models.User, bool, error)
	LoginWithApple(identityToken string, firstName string, lastName string) (string, *models.User, bool, error)

	LikeFood(userID, foodID primitive.ObjectID) (bool, error)
	UnlikeFood(userID, foodID primitive.ObjectID) (bool, error)
	GetLikedFoodIDs(userID primitive.ObjectID) ([]primitive.ObjectID, error)
}

type userService struct {
	userRepo repositories.UserRepository
	foodRepo repositories.FoodRepository
}

func NewUserService(userRepo repositories.UserRepository, foodRepo repositories.FoodRepository) UserService {
	return &userService{userRepo: userRepo, foodRepo: foodRepo}
}

func (s *userService) SignUpWithEmail(input models.SignUpInput) (error) {
	_, err := s.userRepo.FindByEmail(input.Email)
	if err == nil {
		log.Println("Email" + input.Email + "already exists")
		return errors.New("user already exists")
	}

	hashedPasswordBytes, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	hashedPassword := string(hashedPasswordBytes)

	newUser := &models.User{
		ID: primitive.NewObjectID(),
		UserName: input.UserName,
		Email: &input.Email,
		Password: &hashedPassword,
		LoginMethod: models.LoginMethodEmail,
		IsVerified: false,
		Day: 1,
		LikedFoodIDs: make([]primitive.ObjectID, 0),
		CreatedAt: time.Now(),
	}

	err = s.userRepo.Save(newUser)
	if err != nil {
		return err
	}

	return nil
}

func (s *userService) LoginWithEmail(input models.LoginInput) (string, *models.User, error) {
	user, err := s.userRepo.FindByEmail(input.Email)
	if err != nil {
		return "", nil, err // Invalid email
	}

	if user.Password == nil {
		return "", nil, errors.New("this account uses social login")
	}

	err = bcrypt.CompareHashAndPassword([]byte(*user.Password), []byte(input.Password))
	if err != nil {
		return "", nil, err // Invalid password
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID.Hex(),
		"exp": time.Now().Add(time.Hour * 24 * 7).Unix(),
	})
	signedToken, err := token.SignedString([]byte(config.AppConfig.JWTSecret))

	return signedToken, user, err
}

func (s *userService) upsertSocialUser(method, socialID, email, name string) (string, *models.User, bool, error) {
	user, err := s.userRepo.FindBySocialInfo(method, socialID)
	isNew := false

	if err == mongo.ErrNoDocuments {
		isNew = true
		user = &models.User{
			ID: primitive.NewObjectID(),
			UserName: name,
			Email: &email,
			LoginMethod: method,
			SocialID: socialID,
			IsVerified: true,
			Day: 1,
			LikedFoodIDs: make([]primitive.ObjectID, 0),
			CreatedAt: time.Now(),
		}

		if err := s.userRepo.Save(user); err != nil {
			return "", nil, false, err
		}
	} else if err != nil {
		return "", nil, false, err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID.Hex(),
		"exp": time.Now().Add(time.Hour * 24 * 7).Unix(),
	})
	signedToken, err := token.SignedString([]byte(config.AppConfig.JWTSecret))

	return signedToken, user, isNew, err
}

func (s *userService) LoginWithGoogle(idToken string) (string, *models.User, bool, error) {
	webClientID := config.AppConfig.GoogleWebClientID
	payload, err := idtoken.Validate(context.Background(), idToken, webClientID)
	if err != nil {
		log.Println("Google ID token validation error:", err)
		return "", nil, false, errors.New("invalid Google ID token")
	}

	socialID := payload.Claims["sub"].(string)
	email := payload.Claims["email"].(string)
	userName := payload.Claims["name"].(string)

	return s.upsertSocialUser(models.LoginMethodGoogle, socialID, email, userName)
}

func (s *userService) LoginWithKakao(accessToken string) (string, *models.User, bool, error) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", "https://kapi.kakao.com/v2/user/me", nil)
	req.Header.Set("Authorization", "Bearer " + accessToken)

	resp, err := client.Do(req)
	if (err != nil || resp.StatusCode != http.StatusOK) {
		log.Println("Kakao API request error:", err)
		return "", nil, false, errors.New("invalid Kakao access token")
	}
	defer resp.Body.Close()

	var kakaoRes struct {
		ID int64 `json:"id"`
		KakaoAccount struct {
			Email string `json:"email"`
			Profile struct {
				Nickname string `json:"nickname"`
			} `json:"profile"`
		} `json:"kakao_account"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&kakaoRes); err != nil {
		return "", nil, false, errors.New("failed to decode Kakao response")
	}

	socialID := strconv.FormatInt(kakaoRes.ID, 10)
	email := kakaoRes.KakaoAccount.Email
	userName := kakaoRes.KakaoAccount.Profile.Nickname

	return s.upsertSocialUser(models.LoginMethodKakao, socialID, email, userName)
}

func (s *userService) verifyAppleToken(identityToken string, clientID string) (jwt.MapClaims, error) {
	fmt.Println("Starting apple token verification. ClientID:", clientID)

	resp, err := http.Get("https://appleid.apple.com/auth/keys")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("apple api returned status: " + resp.Status)
	}

	var appleKeys AppleKeys
	if err := json.NewDecoder(resp.Body).Decode(&appleKeys); err != nil {
		return nil, err
	}

	token, err := jwt.Parse(identityToken, func(token *jwt.Token) (interface{}, error) {
		kid := token.Header["kid"].(string)
		for _, key := range appleKeys.Keys {
			if key.Kid == kid {
				nBytes, _ := base64.RawURLEncoding.DecodeString(key.N)
				eBytes, _ := base64.RawURLEncoding.DecodeString(key.E)
				var e int
				for _, b := range eBytes {
					e = e<<8 + int(b)
				}
				pubKey := &rsa.PublicKey{
					N: new(big.Int).SetBytes(nBytes),
					E: e,
				}
				return pubKey, nil
			}
		}
		return nil, errors.New("public key not found")
	})

	if err != nil || !token.Valid {
		fmt.Println("Token parsing error:", err)
		return nil, errors.New("invalid Apple identity token")
	}

	claims := token.Claims.(jwt.MapClaims)
	if claims["iss"] != "https://appleid.apple.com" {
		return nil, errors.New("invalid issuer")
	}
	if claims["aud"] != clientID {
		return nil, errors.New("invalid audience")
	}

	fmt.Println("Successfully parsed token.")
	return claims, nil
}

func (s *userService) LoginWithApple(identityToken string, firstName string, lastName string) (string, *models.User, bool, error) {
	clientID := config.AppConfig.AppleBundleID
	claims, err := s.verifyAppleToken(identityToken, clientID)
	if err != nil {
		fmt.Println("Error while verifying id token:", err)
		return "", nil, false, err
	}

	socialID := claims["sub"].(string)
	email, ok := claims["email"].(string)
	if !ok { email = claims["sub"].(string) + "@apple-user.com" }
	userName := firstName + " " + lastName
	if (userName == " ") { userName = "Apple User" }

	return s.upsertSocialUser(models.LoginMethodApple, socialID, email, userName)
}

func (s *userService) LikeFood(userID, foodID primitive.ObjectID) (bool, error) {
	wasAdded, err := s.userRepo.AddLikedFood(userID, foodID)
	if err != nil {
		return false, err
	}

	return wasAdded, nil
}

func (s *userService) UnlikeFood(userID, foodID primitive.ObjectID) (bool, error) {
	wasRemoved, err := s.userRepo.RemoveLikedFood(userID, foodID)
	if err != nil {
		return false, err
	}

	return wasRemoved, nil
}

func (s *userService) GetLikedFoodIDs(userID primitive.ObjectID) ([]primitive.ObjectID, error) {
	return s.userRepo.GetLikedFoodIDs(userID)
}