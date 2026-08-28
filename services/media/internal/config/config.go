package config

import (
	"os"
	"strings"
	"time"

	"github.com/pafthang/arcanum/pkg/svcutil"
	"github.com/pafthang/arcanum/services/media/internal/objectstore"
)

// Config is the media service configuration.
type Config struct {
	MaxBytes    int
	SignTTL     time.Duration
	SignSecret  string
	PublicBase  string
	S3Endpoint  string
	S3Region    string
	S3Bucket    string
	S3AccessKey string
	S3SecretKey string
	S3PathStyle bool
}

// FromEnv loads config from environment.
func FromEnv() Config {
	c := Config{
		MaxBytes:    1 << 20,
		SignTTL:     15 * time.Minute,
		SignSecret:  env("MEDIA_SIGN_SECRET", "dev-media-sign"),
		PublicBase:  strings.TrimRight(env("MEDIA_PUBLIC_BASE", ""), "/"),
		S3Endpoint:  env("MEDIA_S3_ENDPOINT", ""),
		S3Region:    env("MEDIA_S3_REGION", "us-east-1"),
		S3Bucket:    env("MEDIA_S3_BUCKET", ""),
		S3AccessKey: env("MEDIA_S3_ACCESS_KEY", ""),
		S3SecretKey: env("MEDIA_S3_SECRET_KEY", ""),
		S3PathStyle: svcutil.EnvBool("MEDIA_S3_PATH_STYLE", true),
	}
	c.MaxBytes = svcutil.EnvInt("MEDIA_MAX_BYTES", c.MaxBytes)
	if c.MaxBytes < 1024 {
		c.MaxBytes = 1024
	}
	if d := svcutil.EnvDuration("MEDIA_SIGN_TTL", c.SignTTL); d > 0 {
		c.SignTTL = d
	}
	return c
}

func env(key, def string) string {
	s := strings.TrimSpace(os.Getenv(key))
	if s == "" {
		return def
	}
	return s
}

// Backend builds the object store from config.
func (c Config) Backend(dataDir string) objectstore.Backend {
	if strings.TrimSpace(c.S3Bucket) == "" {
		return objectstore.NewFS(dataDir + "/blobs")
	}
	return objectstore.NewS3(c.S3Endpoint, c.S3Region, c.S3Bucket, c.S3AccessKey, c.S3SecretKey, c.S3PathStyle)
}
