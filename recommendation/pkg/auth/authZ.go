package auth

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type AuthZMiddlewareConfig struct {
	RolesClaimName string
	DevMode        bool
}

func NewAuthZMiddleware(config AuthZMiddlewareConfig, requiredRoles []string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if config.DevMode {
			log.Printf("Dev Mode: Toegang toegestaan voor pad: %s", r.URL.Path)
			next.ServeHTTP(w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			log.Printf("Autorisatie header ontbreekt voor pad: %s", r.URL.Path)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			log.Printf("Ongeldig Autorisatie header formaat voor pad: %s", r.URL.Path)
			http.Error(w, "Ongeldig token formaat", http.StatusUnauthorized)
			return
		}

		token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
		if err != nil {
			log.Printf("Fout bij het parsen van JWT (ongeverifieerd) voor pad %s: %v", r.URL.Path, err)
			http.Error(w, "Ongeldig token formaat of claims", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			log.Printf("Ongeldige JWT claims structuur voor pad: %s", r.URL.Path)
			http.Error(w, "Ongeldige token claims", http.StatusUnauthorized)
			return
		}

		userRoles := []string{}
		if rolesClaim, found := claims[config.RolesClaimName]; found {
			if realmAccessMap, isMap := rolesClaim.(map[string]interface{}); isMap {
				if rolesSlice, hasRoles := realmAccessMap["roles"].([]interface{}); hasRoles {
					for _, role := range rolesSlice {
						if rStr, isStr := role.(string); isStr {
							userRoles = append(userRoles, rStr)
						}
					}
				}
			} else if rolesSlice, isSlice := rolesClaim.([]interface{}); isSlice {
				for _, role := range rolesSlice {
					if rStr, isStr := role.(string); isStr {
						userRoles = append(userRoles, rStr)
					}
				}
			}
		}

		isAuthorized := false
		if len(requiredRoles) == 0 {
			isAuthorized = true
		} else {
			for _, requiredRole := range requiredRoles {
				for _, userRole := range userRoles {
					if userRole == requiredRole {
						isAuthorized = true
						break
					}
				}
				if isAuthorized {
					break
				}
			}
		}

		if !isAuthorized {
			log.Printf("Gebruiker (rollen: %v) ongeautoriseerd voor pad: %s. Vereiste rollen: %v", userRoles, r.URL.Path, requiredRoles)
			http.Error(w, "Verboden", http.StatusForbidden)
			return
		}

		log.Printf("Gebruiker (ID: %s, Rollen: %v) geautoriseerd voor pad: %s", claims["sub"], userRoles, r.URL.Path)

		r.Header.Set("X-User-ID", fmt.Sprintf("%v", claims["sub"]))
		r.Header.Set("X-User-Roles", strings.Join(userRoles, ","))
		next.ServeHTTP(w, r)
	})
}
