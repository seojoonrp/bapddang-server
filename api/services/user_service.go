// api/services/user_service.go

package services

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/seojoonrp/bapddang-server/api/repositories"
	"github.com/seojoonrp/bapddang-server/config"
	"github.com/seojoonrp/bapddang-server/models"
	"github.com/seojoonrp/bapddang-server/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/api/idtoken"
)

type AppleKey struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type AppleKeys struct {
	Keys []AppleKey `json:"keys"`
}

type UserService interface {
	SignUp(input models.SignUpInput) (*models.User, error)
	Login(input models.LoginInput) (string, error)
	LoginWithGoogle(idToken string) (bool, string, error)
	LoginWithKakao(accessToken string) (bool, string, error)
	LoginWithApple(identityToken string, firstName string, lastName string) (bool, string, error)

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

func (s *userService) SignUp(input models.SignUpInput) (*models.User, error) {
	_, err := s.userRepo.FindByUsername(input.Username)
	if err == nil {
		return nil, errors.New("user already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	newUser := &models.User{
		ID:           primitive.NewObjectID(),
		Username:     input.Username,
		Password:     string(hashedPassword),
		LoginMethod:  models.LoginMethodEmail,
		Day:          1,
		LikedFoodIDs: make([]primitive.ObjectID, 0),
		CreatedAt:    time.Now(),
	}

	err = s.userRepo.Save(newUser)
	if err != nil {
		return nil, err
	}

	return newUser, nil
}

func (s *userService) Login(input models.LoginInput) (string, error) {
	user, err := s.userRepo.FindByUsername(input.Username)
	if err != nil {
		return "", err // Invalid username
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password))
	if err != nil {
		return "", err // Invalid password
	}

	return utils.GenerateToken(user.ID.Hex())
}

func (s *userService) LoginWithGoogle(idToken string) (bool, string, error) {
	webClientID := config.AppConfig.GoogleWebClientID

	payload, err := idtoken.Validate(context.Background(), idToken, webClientID)
	if err != nil {
		return false, "", errors.New("invalid Google ID token")
	}

	email := payload.Claims["email"].(string)
	userName := payload.Claims["name"].(string)
	isNew := false

	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		isNew = true
		user = &models.User{
			ID:           primitive.NewObjectID(),
			Username:     userName,
			Email:        email,
			Password:     "",
			LoginMethod:  models.LoginMethodGoogle,
			Day:          1,
			LikedFoodIDs: make([]primitive.ObjectID, 0),
			CreatedAt:    time.Now(),
		}

		if err := s.userRepo.Save(user); err != nil {
			return false, "", err
		}
	}

	signedToken, err := utils.GenerateToken(user.ID.Hex())
	return isNew, signedToken, err
}

func (s *userService) LoginWithKakao(accessToken string) (bool, string, error) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", "https://kapi.kakao.com/v2/user/me", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return false, "", errors.New("invalid Kakao access token")
	}
	defer resp.Body.Close()

	var kakaoRes struct {
		ID           int64 `json:"id"`
		KakaoAccount struct {
			Email   string `json:"email"`
			Profile struct {
				Nickname string `json:"nickname"`
			} `json:"profile"`
		} `json:"kakao_account"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&kakaoRes); err != nil {
		return false, "", errors.New("failed to decode Kakao response")
	}

	email := kakaoRes.KakaoAccount.Email
	isNew := false

	user, err := s.userRepo.FindByEmail(email)

	if err != nil {
		isNew = true
		user = &models.User{
			ID:           primitive.NewObjectID(),
			Username:     kakaoRes.KakaoAccount.Profile.Nickname,
			Email:        email,
			Password:     "",
			LoginMethod:  models.LoginMethodKakao,
			Day:          1,
			LikedFoodIDs: make([]primitive.ObjectID, 0),
			CreatedAt:    time.Now(),
		}

		if err := s.userRepo.Save(user); err != nil {
			return false, "", err
		}
	}

	signedToken, err := utils.GenerateToken(user.ID.Hex())
	return isNew, signedToken, err
}

func (s *userService) verifyAppleToken(identityToken string, clientID string) (jwt.MapClaims, error) {
	fmt.Println("Starting apple token verification. ClientID:", clientID)

	resp, err := http.Get("https://appleid.apple.com/auth/keys")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var appleKeys AppleKeys
	json.NewDecoder(resp.Body).Decode(&appleKeys)

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

func (s *userService) LoginWithApple(identityToken string, firstName string, lastName string) (bool, string, error) {
	clientID := config.AppConfig.AppleBundleID
	claims, err := s.verifyAppleToken(identityToken, clientID)
	if err != nil {
		fmt.Println("Error while verifying id token:", err)
		return false, "", err
	}

	email, ok := claims["email"].(string)
	if !ok {
		email = claims["sub"].(string) + "@apple-user.com"
	}
	fmt.Println("Apple login for email:", email)

	userName := firstName + " " + lastName
	if userName == " " {
		userName = "Apple User"
	}

	isNew := false
	user, err := s.userRepo.FindByEmail(email)

	if err != nil {
		isNew = true
		user = &models.User{
			ID:           primitive.NewObjectID(),
			Username:     userName,
			Email:        email,
			Password:     "",
			LoginMethod:  models.LoginMethodApple,
			Day:          1,
			LikedFoodIDs: make([]primitive.ObjectID, 0),
			CreatedAt:    time.Now(),
		}

		if err := s.userRepo.Save(user); err != nil {
			return false, "", err
		}
	}

	signedToken, err := utils.GenerateToken(user.ID.Hex())
	return isNew, signedToken, err
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
