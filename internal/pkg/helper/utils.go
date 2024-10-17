package helper

import (
	"crypto/sha1"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/constants"
	"github.com/spf13/viper"
)

const urlRegex string = `^((ftp|http|https):\/\/)?(\S+(:\S*)?@)?((([1-9]\d?|1\d\d|2[01]\d|22[0-3])(\.(1?\d{1,2}|2[0-4]\d|25[0-5])){2}(?:\.([0-9]\d?|1\d\d|2[0-4]\d|25[0-4]))|(((([a-z\x{00a1}-\x{ffff}0-9]+-?-?_?)*[a-z\x{00a1}-\x{ffff}0-9]+)\.)?)?(([a-z\x{00a1}-\x{ffff}0-9]+-?-?_?)*[a-z\x{00a1}-\x{ffff}0-9]+)(?:\.([a-z\x{00a1}-\x{ffff}]{2,}))?)|localhost)(:(\d{1,5}))?((\/|\?|#)[^\s]*)?$`

// Base64EncodedTLS creates a TLS configuration using base64-encoded certificate, key, and CA files.
// It returns a pointer to tls.Config, which can be used in secure communication setups.
func Base64EncodedTLS(ca, cert, key string) (*tls.Config, error) {
	certPem, err := base64.StdEncoding.DecodeString(cert)
	if err != nil {
		return nil, fmt.Errorf("unabled to decode cert: %v", err)
	}

	keyPem, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return nil, fmt.Errorf("unable to decode key: %v", err)
	}

	certs, err := tls.X509KeyPair(certPem, keyPem)
	if err != nil {
		return nil, fmt.Errorf("unable to load key pair: %v", err)
	}

	caPem, err := base64.StdEncoding.DecodeString(ca)
	if err != nil {
		return nil, fmt.Errorf("unabled to load CA: %v", err)
	}

	caPool := x509.NewCertPool()
	caPool.AppendCertsFromPEM(caPem)

	return &tls.Config{
		Certificates: []tls.Certificate{certs},
		RootCAs:      caPool,
	}, nil
}

// RedisPrefixes generates a prefixed key for Redis based on the input parameters.
// The function constructs the prefixed key by concatenating the prefixKey, expectedPrefixKey (if provided).
// and the main key, seperated by appropiate delimiters. The constructed key is then returned.
// If expectedPrefixKey is provided, the redisId is added as a prefix before it.
func RedisPrefixes(key, prefixKey, expectedPrefixKey, redisId string) string {
	prefix := prefixKey + ":" + key

	if expectedPrefixKey != "" {
		prefix = redisId + "_" + expectedPrefixKey + ":" + key
	}

	return prefix
}

// ViperReader reads a configuration value from a given string and unmarshals it into a provided variable
// The method reads the configuration data from the provided string using Viper and unmarshals it into the variable specified by rawVal.
func ViperReader(key, value string, rawVal any) error {
	if err := viper.ReadConfig(strings.NewReader(value)); err != nil {
		return err
	}
	if err := viper.UnmarshalKey(key, rawVal); err != nil {
		return err
	}
	return nil
}

func StringSliceChecker(s []string) []string {
	var r []string

	for _, str := range s {
		if str != "" {
			r = append(r, str)
		}
	}

	return r
}

func StringBuilder(params ...string) string {
	var sb strings.Builder

	for _, param := range params {
		sb.WriteString(param)
	}

	return sb.String()
}

func BbbUrlBuilder(path string, salt string, params ...string) string {
	h := sha1.New()
	var sb strings.Builder

	sb.WriteString(path)
	for _, param := range params {
		sb.WriteString(param)
	}

	a := func() string {
		sb.WriteString(salt)
		salted := sb.String()
		h.Write([]byte(salted))
		hash := hex.EncodeToString(h.Sum(nil))
		return hash
	}

	sx := a()

	var sb2 strings.Builder
	sb2.WriteString(path + "?")
	for _, param := range params {
		sb2.WriteString(param)
	}
	sb2.WriteString("&checksum=" + sx)

	return sb2.String()
}

var urlRegexPattern = regexp.MustCompile(urlRegex)

func UrlChecker(url string) bool {
	return urlRegexPattern.MatchString(url)
}

func IsValidURL(rawURL string) bool {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return false
	}

	return parsedURL.Scheme != "" && parsedURL.Host != ""
}

func PelangganCounter(i int64) (string, error) {
	result := fmt.Sprintf("%d%05d", (i/100000)+1, (i % 100000))

	limit := 997999

	if resultInt, err := strconv.Atoi(result); err == nil && resultInt > limit {
		return "", errors.New("result exceeds the limit of 997999")
	}

	return result, nil
}

func AplikatorCounter(i int, level int32) (string, int32, error) {
	switch {
	case level == constants.ManagerLevel:
		return "", 0, fmt.Errorf("there should be only 1 manager in aplikator system")
	case level == constants.InventoryLevel:
		if i+999800 > 999850 {
			return "", 0, fmt.Errorf("inventory users already exceeding the limit, which is 50. please consider editing the existing unused user")
		}
		result := fmt.Sprintf("%06d", i+999800)
		return result, constants.InventoryLevel, nil
	case level == constants.AkuntingLevel:
		if i+999850 > 999900 {
			return "", 0, fmt.Errorf("akunting users already exceeding the limit, which is 50. please consider editing the existing unused user")
		}
		result := fmt.Sprintf("%06d", i+999850)
		return result, constants.InventoryLevel, nil
	default:
		return "", 0, fmt.Errorf("the provided level might not exist in the system")
	}
}

func CheckSystemLevel(level int) error {
	switch level {
	case int(constants.StokistLevel), int(constants.PelangganLevel), int(constants.AplikatorLevel), int(constants.SystemLevel):
		return nil
	default:
		return fmt.Errorf("invalid system level: %v", level)
	}
}

func CheckSystemLevelLabel(label string) error {
	switch label {
	case constants.Stokist, constants.Pelanggan, constants.Aplikator, constants.System:
		return nil
	default:
		return fmt.Errorf("invalid system level label: %v", label)
	}
}

func CheckAplikatorLevel(level int) error {
	switch level {
	case int(constants.InventoryLevel), int(constants.AkuntingLevel):
		return nil
	default:
		return fmt.Errorf("invalid aplikator level: %v", level)
	}
}

func CheckAplikatorLevelLabel(label string) error {
	switch label {
	case constants.Inventory, constants.Akunting:
		return nil
	default:
		return fmt.Errorf("invalid aplikator level label: %v", label)
	}
}

func IsValidImageContentType(contentType string) bool {
	contentType = strings.ToLower(contentType)
	return contentType == "image/jpeg" || contentType == "image/png" || contentType == "image/webp"
}

func IsValidPlatform(platformKey, mobilePlatform, webPlatform string) bool {
	return platformKey == webPlatform || platformKey == mobilePlatform
}
