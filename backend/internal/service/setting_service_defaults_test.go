//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

type settingDefaultsRepoStub struct {
	getValueErr error
	updates     map[string]string
}

func (s *settingDefaultsRepoStub) Get(ctx context.Context, key string) (*Setting, error) {
	panic("unexpected Get call")
}

func (s *settingDefaultsRepoStub) GetValue(ctx context.Context, key string) (string, error) {
	if s.getValueErr != nil {
		return "", s.getValueErr
	}
	return "existing", nil
}

func (s *settingDefaultsRepoStub) Set(ctx context.Context, key, value string) error {
	panic("unexpected Set call")
}

func (s *settingDefaultsRepoStub) GetMultiple(ctx context.Context, keys []string) (map[string]string, error) {
	panic("unexpected GetMultiple call")
}

func (s *settingDefaultsRepoStub) SetMultiple(ctx context.Context, settings map[string]string) error {
	s.updates = make(map[string]string, len(settings))
	for key, value := range settings {
		s.updates[key] = value
	}
	return nil
}

func (s *settingDefaultsRepoStub) GetAll(ctx context.Context) (map[string]string, error) {
	panic("unexpected GetAll call")
}

func (s *settingDefaultsRepoStub) Delete(ctx context.Context, key string) error {
	panic("unexpected Delete call")
}

func TestSettingService_InitializeDefaultSettings_IncludesLargeTablePageSizeOption(t *testing.T) {
	repo := &settingDefaultsRepoStub{getValueErr: ErrSettingNotFound}
	svc := NewSettingService(repo, &config.Config{})

	err := svc.InitializeDefaultSettings(context.Background())

	require.NoError(t, err)
	require.Equal(t, "[10,20,50,100,1000]", repo.updates[SettingKeyTablePageSizeOptions])
}
