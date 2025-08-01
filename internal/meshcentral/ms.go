package meshcentral

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"strings"
	"time"

	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/gorilla/websocket"
)

type MeshResponse struct {
	Action string `json:"action"`
	Meshes []struct {
		Name string `json:"name"`
		ID   string `json:"_id"`
	} `json:"meshes"`
}

func getMeshDeviceGroupID(ctx context.Context, uri, deviceGroup string) (string, error) {
	conn, _, err := websocket.DefaultDialer.Dial(uri, nil)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	err = conn.WriteMessage(websocket.TextMessage, []byte(`{"action": "meshes", "responseid": "meshctrl"}`))
	if err != nil {
		return "", err
	}

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			return "", err
		}

		var response MeshResponse
		if err := json.Unmarshal(message, &response); err != nil {
			return "", err
		}

		if response.Action == "meshes" {
			for _, mesh := range response.Meshes {
				if mesh.Name == deviceGroup {
					return mesh.ID[len("mesh//"):], nil
				}
			}
			return "", errors.New("device group not found")
		}
	}
}

func GetMeshDeviceGroupId(uri string, deviceGroup string) (id string, err error) {
	ctx := context.Background()

	id, err = getMeshDeviceGroupID(ctx, uri, deviceGroup)
	if err != nil {
		return "", err
	}

	return id, nil
}

func formatUserID(user, domain string) string {
	return "user/" + domain + "/" + user
}

func getAuthToken(user, key, domain string) (string, error) {
	keyBytes, err := hex.DecodeString(key)
	if err != nil {
		return "", err
	}
	if len(keyBytes) < 32 {
		return "", errors.New("key length must be at least 32 bytes")
	}
	key1 := keyBytes[:32]

	msg := fmt.Sprintf(`{"userid":"%s", "domainid":"%s", "time":%d}`,
		formatUserID(user, domain), domain, time.Now().Unix())

	//fmt.Println("msg: ", msg)
	//iv, err := hex.DecodeString("000000000000000000000000")
	iv := make([]byte, 12)
	_, err = rand.Read(iv)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key1)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Encrypt and generate authentication tag
	ciphertext := aesGCM.Seal(nil, iv, []byte(msg), nil)
	tag := ciphertext[len(ciphertext)-aesGCM.Overhead():]
	ciphertext = ciphertext[:len(ciphertext)-aesGCM.Overhead()]

	// Concatenate IV, tag, and ciphertext
	data := append(iv, tag...)
	data = append(data, ciphertext...)

	// Base64 encode and replace characters to match Python's altchars "@$"
	encoded := base64.StdEncoding.EncodeToString(data)
	encoded = strings.ReplaceAll(encoded, "/", "$")
	encoded = strings.ReplaceAll(encoded, "+", "@")

	return encoded, nil
}

func GetMeshWsUrl(uri string, user string, token string) (string, error) {
	newToken, err := getAuthToken(user, token, "")
	if err != nil {
		return "", err
	}

	u, _ := url.Parse(uri)
	return "wss://" + u.Host + "/control.ashx?auth=" + newToken, nil
}
