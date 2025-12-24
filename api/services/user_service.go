// api/services/user_service.go

package services

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/seojoonrp/bapddang-server/api/repositories"
	"github.com/seojoonrp/bapddang-server/config"
	"github.com/seojoonrp/bapddang-server/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
	SignUp(input models.SignUpInput) (*models.User, error)
	Login(input models.LoginInput) (string, error)
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

func (s *userService) SignUp(input models.SignUpInput) (*models.User, error) {
	_, err := s.userRepo.FindByEmail(input.Email)
	if err == nil {
		return nil, errors.New("user already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	newUser := &models.User{
		ID:           primitive.NewObjectID(),
		UserName:     input.UserName,
		Email:        input.Email,
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
	user, err := s.userRepo.FindByEmail(input.Email)
	if err != nil {
		return "", err // Invalid email
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password))
	if err != nil {
		return "", err // Invalid password
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID.Hex(),
		"exp": time.Now().Add(time.Hour * 24 * 7).Unix(),
	})

	signedToken, err := token.SignedString([]byte(config.AppConfig.JWTSecret))
	return signedToken, err
}

func (s *userService) LoginWithGoogle(idToken string) (string, *models.User, bool, error) {
	webClientID := config.AppConfig.GoogleWebClientID

	payload, err := idtoken.Validate(context.Background(), idToken, webClientID)
	if err != nil {
		return "", nil, false, errors.New("invalid Google ID token")
	}

	email := payload.Claims["email"].(string)
	userName := payload.Claims["name"].(string)
	isNew := false

	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		isNew = true
		user = &models.User{
			ID: primitive.NewObjectID(),
			UserName: userName,
			Email:  email,
			Password: "",
			LoginMethod: models.LoginMethodGoogle,
			Day: 1,
			LikedFoodIDs: make([]primitive.ObjectID, 0),
			CreatedAt: time.Now(),
		}

		if err := s.userRepo.Save(user); err != nil {
			return "", nil, false, err
		}
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID.Hex(),
		"exp": time.Now().Add(time.Hour * 24 * 7).Unix(),
	})

	signedToken, err := token.SignedString([]byte(config.AppConfig.JWTSecret))
	return signedToken, user, isNew, err
}

func (s *userService) LoginWithKakao(accessToken string) (string, *models.User, bool, error) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", "https://kapi.kakao.com/v2/user/me", nil)
	req.Header.Set("Authorization", "Bearer " + accessToken)

	resp, err := client.Do(req)
	if (err != nil || resp.StatusCode != http.StatusOK) {
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

	email := kakaoRes.KakaoAccount.Email
	isNew := false

	user, err := s.userRepo.FindByEmail(email)

	if err != nil {
		isNew = true
		user = &models.User{
			ID: primitive.NewObjectID(),
			UserName: kakaoRes.KakaoAccount.Profile.Nickname,
			Email: email,
			Password: "",
			LoginMethod: models.LoginMethodKakao,
			Day: 1,
			LikedFoodIDs: make([]primitive.ObjectID, 0),
			CreatedAt: time.Now(),
		}

		if err := s.userRepo.Save(user); err != nil {
			return "", nil, false, err
		}
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID.Hex(),
		"exp": time.Now().Add(time.Hour * 24 * 7).Unix(),
	})

	signedToken, err := token.SignedString([]byte(config.AppConfig.JWTSecret))
	return signedToken, user, isNew, err
}

func (s *userService) verifyAppleToken(identityToken string, clientID string) (jwt.MapClaims, error) {
	resp, _ := http.Get("https://appleid.apple.com/auth/keys")
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
		return nil, errors.New("invalid Apple identity token")
	}

	claims := token.Claims.(jwt.MapClaims)
	if claims["iss"] != "https://appleid.apple.com" {
		return nil, errors.New("invalid issuer")
	}
	if claims["aud"] != clientID {
		return nil, errors.New("invalid audience")
	}

	return claims, nil
}

func (s *userService) LoginWithApple(identityToken string, firstName string, lastName string) (string, *models.User, bool, error) {
	clientID := config.AppConfig.AppleBundleID
	claims, err := s.verifyAppleToken(identityToken, clientID)
	if err != nil {
		return "", nil, false, err
	}

	email := claims["email"].(string)
	userName := firstName + " " + lastName
	if (userName == " ") {
		userName = "Apple User"
	}

	isNew := false
	user, err := s.userRepo.FindByEmail(email)

	if err != nil {
		isNew = true
		user = &models.User{
			ID:           primitive.NewObjectID(),
			UserName:     userName,
			Email:        email,
			Password:     "",
			LoginMethod:  models.LoginMethodApple,
			Day:          1,
			LikedFoodIDs: make([]primitive.ObjectID, 0),
			CreatedAt:    time.Now(),
		}

		if err := s.userRepo.Save(user); err != nil {
			return "", nil, false, err
		}
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID.Hex(),
		"exp": time.Now().Add(time.Hour * 24 * 7).Unix(),
	})

	signedToken, err := token.SignedString([]byte(config.AppConfig.JWTSecret))
	return signedToken, user, isNew, err
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