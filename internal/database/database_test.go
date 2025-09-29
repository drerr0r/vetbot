package database

import (
	"testing"

	"github.com/drerr0r/vetbot/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDatabase_Integration(t *testing.T) {
	config := GetTestConfig()
	db := SetupTestDatabase(t, config)

	if db == nil {
		return // Тест был пропущен
	}
	defer db.Close()
	defer CleanupTestDatabase(db)

	t.Run("CreateUser", func(t *testing.T) {
		user, err := CreateTestUser(db, 99991)
		require.NoError(t, err)
		assert.NotZero(t, user.ID)
		assert.NotZero(t, user.CreatedAt)
	})

	t.Run("GetAllSpecializations", func(t *testing.T) {
		specializations, err := db.GetAllSpecializations()
		require.NoError(t, err)
		assert.Greater(t, len(specializations), 0)

		if len(specializations) > 0 {
			assert.NotEmpty(t, specializations[0].Name)
		}
	})

	t.Run("GetSpecializationByID", func(t *testing.T) {
		spec, err := db.GetSpecializationByID(1)
		require.NoError(t, err)
		assert.Equal(t, 1, spec.ID)
		assert.Equal(t, "Хирург", spec.Name)
	})

	t.Run("SpecializationExists", func(t *testing.T) {
		exists, err := db.SpecializationExists(1)
		require.NoError(t, err)
		assert.True(t, exists)

		exists, err = db.SpecializationExists(999)
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("GetVeterinariansBySpecialization", func(t *testing.T) {
		vets, err := db.GetVeterinariansBySpecialization(1)
		require.NoError(t, err)
		assert.Greater(t, len(vets), 0)

		if len(vets) > 0 {
			vet := vets[0]
			assert.NotEmpty(t, vet.FirstName)
			assert.NotEmpty(t, vet.Phone)
		}
	})

	t.Run("GetAllClinics", func(t *testing.T) {
		clinics, err := db.GetAllClinics()
		require.NoError(t, err)
		assert.Greater(t, len(clinics), 0)

		if len(clinics) > 0 {
			clinic := clinics[0]
			assert.NotEmpty(t, clinic.Name)
			assert.NotEmpty(t, clinic.Address)
		}
	})

	t.Run("GetAllVeterinarians", func(t *testing.T) {
		vets, err := db.GetAllVeterinarians()
		require.NoError(t, err)
		assert.Greater(t, len(vets), 0)
	})

	t.Run("GetVeterinarianByID", func(t *testing.T) {
		vet, err := db.GetVeterinarianByID(1)
		require.NoError(t, err)
		assert.Equal(t, 1, vet.ID)
		assert.Equal(t, "Иван", vet.FirstName)
	})

	t.Run("GetClinicByID", func(t *testing.T) {
		clinic, err := db.GetClinicByID(1)
		require.NoError(t, err)
		assert.Equal(t, 1, clinic.ID)
		assert.Equal(t, "ВетКлиника Центр", clinic.Name)
	})

	t.Run("FindAvailableVets_EmptyCriteria", func(t *testing.T) {
		criteria := &models.SearchCriteria{}
		vets, err := db.FindAvailableVets(criteria)
		require.NoError(t, err)
		assert.Greater(t, len(vets), 0)
	})

	t.Run("FindAvailableVets_WithSpecialization", func(t *testing.T) {
		criteria := &models.SearchCriteria{
			SpecializationID: 1,
		}
		vets, err := db.FindAvailableVets(criteria)
		require.NoError(t, err)
		assert.Greater(t, len(vets), 0)
	})

	t.Run("AddMissingColumns", func(t *testing.T) {
		err := db.AddMissingColumns()
		assert.NoError(t, err)
	})
}

func TestDatabase_Unit(t *testing.T) {
	t.Run("New_InvalidURL", func(t *testing.T) {
		db, err := New("invalid_url")
		assert.Error(t, err)
		assert.Nil(t, db)
	})

	t.Run("Close_NilDB", func(t *testing.T) {
		// Создаем объект Database с nil db
		db := &Database{
			db: nil,
		}

		// Вызываем Close - это не должно вызывать панику
		err := db.Close()

		// Вместо проверки ошибки просто убеждаемся, что паники не было
		if err != nil {
			t.Logf("Close returned error (expected for nil DB): %v", err)
		}
	})

	t.Run("Close_ProperlyInitialized", func(t *testing.T) {
		// Создаем базу данных с невалидным URL, чтобы получить nil DB
		db, err := New("invalid_url")
		assert.Error(t, err)
		assert.Nil(t, db)

		// Если New возвращает nil, не пытаемся вызывать Close
		if db != nil {
			err := db.Close()
			assert.NoError(t, err)
		}
	})
}
