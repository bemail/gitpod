// Copyright (c) 2022 Gitpod GmbH. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License-AGPL.txt in the project root for license information.

package db

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"time"

	"github.com/gitpod-io/gitpod/common-go/log"
	driver_mysql "github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type ConnectionParams struct {
	User         string
	Password     string
	Host         string
	Database     string
	CustomCACert string
}

func Connect(p ConnectionParams) (*gorm.DB, error) {
	loc, err := time.LoadLocation("UTC")
	if err != nil {
		return nil, fmt.Errorf("Failed to load UT location: %w", err)
	}
	cfg := driver_mysql.Config{
		User:                 p.User,
		Passwd:               p.Password,
		Net:                  "tcp",
		Addr:                 p.Host,
		DBName:               p.Database,
		Loc:                  loc,
		AllowNativePasswords: true,
		ParseTime:            true,
	}

	if p.CustomCACert != "" {
		rootCertPool := x509.NewCertPool()
		if ok := rootCertPool.AppendCertsFromPEM([]byte(p.CustomCACert)); !ok {
			log.Fatal("Failed to append custom DB CA cert.")
		}

		tlsConfigName := "custom"
		err = driver_mysql.RegisterTLSConfig(tlsConfigName, &tls.Config{
			RootCAs: rootCertPool,
		})
		if err != nil {
			return nil, fmt.Errorf("Failed to register custom DB CA cert: %w", err)
		}
		cfg.TLSConfig = tlsConfigName
	}

	// refer to https://github.com/go-sql-driver/mysql#dsn-data-source-name for details
	return gorm.Open(mysql.Open(cfg.FormatDSN()), &gorm.Config{
		Logger: logger.New(log.Log, logger.Config{
			SlowThreshold: 200 * time.Millisecond,
			Colorful:      false,
			LogLevel: (func() logger.LogLevel {
				switch log.Log.Level {
				case logrus.PanicLevel, logrus.FatalLevel, logrus.ErrorLevel:
					return logger.Error
				case logrus.WarnLevel:
					return logger.Warn
				default:
					return logger.Info
				}
			})(),
		}),
	})
}
