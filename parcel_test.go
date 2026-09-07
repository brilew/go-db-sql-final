package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()
	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	number, err := store.Add(parcel)
	require.NoError(t, err)

	require.Greater(t, number, 0)
	// get
	parc, err := store.Get(number)
	require.NoError(t, err)
	assert.Equal(t, parcel.Address, parc.Address)
	assert.Equal(t, parcel.Client, parc.Client)
	assert.Equal(t, parcel.CreatedAt, parc.CreatedAt)
	assert.Equal(t, parcel.Status, parc.Status)
	// delete

	err = store.Delete(number)
	require.NoError(t, err)

	_, err = store.Get(number)
	require.Error(t, err)
	assert.ErrorIs(t, err, sql.ErrNoRows)
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	// add

	parcel := getTestParcel()
	number, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotZero(t, number)
	// set address

	newAddress := "new test address"
	err = store.SetAddress(number, newAddress)
	require.NoError(t, err)

	// check

	parcel, err = store.Get(number)
	require.NoError(t, err)
	assert.Equal(t, newAddress, parcel.Address)
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()
	store := NewParcelStore(db)
	// add

	parcel := Parcel{
		Client:    4567,
		Status:    ParcelStatusRegistered,
		Address:   "Australia, Vellington, Lenin's road 13-22",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	number, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotZero(t, number)
	// set status

	err = store.SetStatus(number, ParcelStatusSent)
	require.NoError(t, err)
	// check

	parcel, err = store.Get(number)
	require.NoError(t, err)
	assert.Equal(t, ParcelStatusSent, parcel.Status)
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}

	// задаём всем посылкам один и тот же идентификатор клиента
	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	// add
	for i := 0; i < len(parcels); i++ {
		id, err := store.Add(parcels[i])
		require.NoError(t, err)
		parcels[i].Number = id
		parcelMap[id] = parcels[i]
	}

	// get by client
	storedParcels, err := store.GetByClient(client)
	require.NoError(t, err)

	assert.Len(t, parcels, len(storedParcels))
	// check it out
	for _, parcel := range storedParcels {

		prcl, ok := parcelMap[parcel.Number]
		require.True(t, ok)
		assert.Equal(t, parcel, prcl)

	}
}
