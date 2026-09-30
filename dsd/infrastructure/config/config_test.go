package config_test

import (
	"os"

	"github.com/Pluslab/cyphonic/dsd/infrastructure/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config", func() {
	Describe("Get", func() {
		AfterEach(func() {
			os.Clearenv()
		})

		Context("when all required environment variables are set", func() {
			It("should return a valid config", func() {
				os.Setenv("UDP_ADDRESS", "udpAddress")
				os.Setenv("UDP_PORT", "udpPort")
				os.Setenv("SERVICE_TYPE", "serviceType")
				os.Setenv("DATABASE_USER", "databaseUser")
				os.Setenv("DATABASE_HOST", "databaseHost")
				os.Setenv("DATABASE_PORT", "databasePort")
				os.Setenv("DATABASE_NAME", "databaseName")
				os.Setenv("DATABASE_SSL_MODE", "databaseSSLMode")
				os.Setenv("DATABASE_TIME_ZONE", "databaseTimeZone")
				os.Setenv("REDIS_ADDRESSES", "redisAddresses")
				os.Setenv("FQDN", "fqdn")
				os.Setenv("PORT", "port")
				os.Setenv("DEBUG_LOG_FILE_PATH", "debugLogFilePath")
				os.Setenv("ERROR_LOG_FILE_PATH", "errorLogFilePath")
				os.Setenv("CACHE_CLUSTER_MODE", "true")
				os.Setenv("DEBUG_MODE", "true")
				os.Setenv("DB_DEBUG_MODE", "true")

				cfg, err := config.Get()
				Expect(err).ToNot(HaveOccurred())
				Expect(cfg.ServiceType).To(Equal("serviceType"))
				Expect(cfg.DatabaseUser).To(Equal("databaseUser"))
				Expect(cfg.DatabaseHost).To(Equal("databaseHost"))
				Expect(cfg.DatabasePort).To(Equal("databasePort"))
				Expect(cfg.DatabaseName).To(Equal("databaseName"))
				Expect(cfg.DatabaseSSLMode).To(Equal("databaseSSLMode"))
				Expect(cfg.DatabaseTimeZone).To(Equal("databaseTimeZone"))
				Expect(cfg.RedisAddresses).To(Equal([]string{"redisAddresses"}))
				Expect(cfg.FQDN).To(Equal("fqdn"))
				Expect(cfg.Port).To(Equal("port"))
				Expect(cfg.DebugLogFilePath).To(Equal("debugLogFilePath"))
				Expect(cfg.ErrorLogFilePath).To(Equal("errorLogFilePath"))
				Expect(cfg.CacheClusterMode).To(BeTrue())
				Expect(cfg.DebugMode).To(BeTrue())
				Expect(cfg.DBDebugMode).To(BeTrue())
			})
		})

		Context("when a required environment variable is missing", func() {
			It("should return an error", func() {
				os.Setenv("UDP_ADDRESS", "udpAddress")
				os.Setenv("UDP_PORT", "udpPort")
				os.Setenv("SERVICE_TYPE", "serviceType")
				os.Setenv("DATABASE_USER", "databaseUser")
				os.Setenv("DATABASE_HOST", "databaseHost")
				os.Setenv("DATABASE_PORT", "databasePort")
				os.Setenv("DATABASE_NAME", "databaseName")
				os.Setenv("DATABASE_SSL_MODE", "databaseSSLMode")
				os.Setenv("DATABASE_TIME_ZONE", "databaseTimeZone")
				os.Setenv("REDIS_ADDRESSES", "redisAddresses")
				os.Setenv("FQDN", "fqdn")
				os.Setenv("PORT", "port")
				os.Setenv("DEBUG_LOG_FILE_PATH", "debugLogFilePath")
				os.Setenv("ERROR_LOG_FILE_PATH", "errorLogFilePath")
				os.Setenv("CACHE_CLUSTER_MODE", "true")
				os.Setenv("DEBUG_MODE", "true")

				_, err := config.Get()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("missing required environment variables"))
			})
		})
	})
})
