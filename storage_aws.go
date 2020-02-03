package cryptoengine

import (
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/pkg/errors"
)

// StorageAws is a storage engine that uses AWS Secrets Manager to persist private keys.
type StorageAws struct {
	keys           map[string][]byte
	secretsManager *secretsmanager.SecretsManager
	secretPrefix   string
}

// Read...
func (s *StorageAws) Read(name string) ([]byte, error) {
	if s == nil {
		return nil, nil
	}

	if s.keys == nil {
		s.keys = make(map[string][]byte)
	}

	dat, ok := s.keys[name]
	if !ok {
		secretId := filepath.Join(s.secretPrefix, name)

		res, err := s.secretsManager.GetSecretValue(&secretsmanager.GetSecretValueInput{
			SecretId: aws.String(secretId),
		})
		// Flag whether the secret exists and update needs to be used
		// instead of create.
		if err != nil {
			var awsSecretIDNotFound bool
			if aerr, ok := err.(awserr.Error); ok {
				switch aerr.Code() {
				case secretsmanager.ErrCodeResourceNotFoundException:
					awsSecretIDNotFound = true
				}
			}

			if !awsSecretIDNotFound {
				return nil, errors.WithMessagef(err, "Get secret value for %s failed", secretId)
			}
		} else {
			dat = res.SecretBinary
			s.keys[name] = dat
		}
	}

	return dat, nil
}

// Write...
func (s *StorageAws) Write(name string, dat []byte) error {
	if s == nil {
		return nil
	}

	secretId := filepath.Join(s.secretPrefix, name)

	_, err := s.secretsManager.CreateSecret(&secretsmanager.CreateSecretInput{
		Name:         aws.String(secretId),
		SecretBinary: dat,
	})
	// Flag whether the secret exists and update needs to be used
	// instead of create.
	if err != nil {
		var awsSecretExists bool
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case secretsmanager.ErrCodeResourceExistsException:
				awsSecretExists = true
			}
		}
		if !awsSecretExists {
			return errors.WithMessagef(err, "Put secret value for %s failed", secretId)
		}

		_, err = s.secretsManager.UpdateSecret(&secretsmanager.UpdateSecretInput{
			SecretId:     aws.String(secretId),
			SecretBinary: dat,
		})
		if err != nil {
			return err
		}
	}

	if s.keys == nil {
		s.keys = make(map[string][]byte)
	}
	s.keys[name] = dat

	return nil
}

// Delete...
func (s *StorageAws) Delete(name string) error {
	if s == nil || s.keys == nil {
		return nil
	}

	delete(s.keys, name)

	secretId := filepath.Join(s.secretPrefix, name)

	_, err := s.secretsManager.DeleteSecret(&secretsmanager.DeleteSecretInput{
		SecretId: aws.String(secretId),
	})
	// Flag whether the secret exists and update needs to be used
	// instead of create.
	if err != nil {
		var awsSecretIDNotFound bool
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case secretsmanager.ErrCodeResourceNotFoundException:
				awsSecretIDNotFound = true
			}
		}

		if !awsSecretIDNotFound {
			return errors.WithMessagef(err, "Delete secret value for %s failed", secretId)
		}
	}

	return nil
}

// NewStorageAws implements the interface Storage to support persisting private keys to AWS Secrets Manager.
func NewStorageAws(awsSession *session.Session, awsSecretPrefix string) (*StorageAws, error) {
	if awsSession == nil {
		return nil, errors.New("aws session cannot be nil")
	}

	if awsSecretPrefix == "" {
		return nil, errors.New("aws secret prefix cannot be empty")
	}

	sm := secretsmanager.New(awsSession)

	storage := &StorageAws{
		keys:           make(map[string][]byte),
		secretsManager: sm,
		secretPrefix:   awsSecretPrefix,
	}

	return storage, nil
}
