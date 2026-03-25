package cognito

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
)

func (cp *CognitoProvider) secretHash(username string) string {

	mac := hmac.New(sha256.New, []byte(cp.cfg.ClientSecret))
	mac.Write([]byte(username + cp.cfg.ClientID))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}
