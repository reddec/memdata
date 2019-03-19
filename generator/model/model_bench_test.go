package model

import (
	"sync"
	"testing"
)

type DataReader interface {
	User(id int64) *User
}

type DataWriter interface {
	InsertUser(item *User) *User
	RemoveUser(id int64)
	UpdateUser(item *User) *User
}

type DataReadWriter interface {
	DataReader
	DataWriter
}

type DataReadWriterTx interface {
	DataReader
	DataWriter
	Commit()
	Discard()
}

type DataReaderTx interface {
	DataReader
	ReadUnlock()
}

type DataAction int

const (
	DataActionUnknown DataAction = 0
	DataActionInsert  DataAction = 1
	DataActionUpdate  DataAction = 2
	DataActionDelete  DataAction = 3
)

type DataLogEntity struct {
	// only of log entity should be filled
	User *UserLogEntity
}
type Data interface {
	ReadLock() DataReaderTx
	ReadWriteLock() DataReadWriterTx
}

type DataTxStorage interface {
	GetUser(id int64) *User
	IterateUser(iterator func(id int64, item *User))
	Apply(batch []DataLogEntity)
}

type implData struct {
	sequenceUserId int64
	storage        DataTxStorage
	_tx            sync.RWMutex
	_log           []DataLogEntity
}

func NewData(storage DataTxStorage) Data {
	// restore auto-sequence for User.Id
	var maxIdOfUser int64
	storage.IterateUser(func(id int64, item *User) {
		if id > maxIdOfUser {
			maxIdOfUser = id
		}
	})
	return &implData{storage: storage, sequenceUserId: maxIdOfUser}
}
func DefaultData() Data {
	return NewData(NewMapDataStorage())
}
func (project *implData) ReadLock() DataReaderTx {
	project._tx.RLock()
	return project
}
func (project *implData) ReadWriteLock() DataReadWriterTx {
	project._tx.Lock()
	return project
}
func (project *implData) ReadUnlock() {
	project._tx.RUnlock()
}
func (project *implData) Commit() {
	project.storage.Apply(project._log)
	project.Discard()
}
func (project *implData) Discard() {
	if project._log != nil {
		project._log = project._log[:0]
	}
	project._tx.Unlock()
}
func (project *implData) User(id int64) *User {
	return project.storage.GetUser(id)
}
func (project *implData) NextUserId() int64 {
	project.sequenceUserId++
	return project.sequenceUserId
}
func (project *implData) InsertUser(item *User) *User {
	item.Id = project.NextUserId()
	item._project = project
	project._log = append(project._log, DataLogEntity{User: &UserLogEntity{Id: item.Id, Item: *item, Action: DataActionInsert}})
	return item
}
func (project *implData) UpdateUser(item *User) *User {
	project._log = append(project._log, DataLogEntity{User: &UserLogEntity{Id: item.Id, Item: *item, Action: DataActionUpdate}})
	return item
}
func (project *implData) RemoveUser(id int64) {
	project._log = append(project._log, DataLogEntity{User: &UserLogEntity{Id: id, Action: DataActionDelete}})
}

type memDataMapStorage struct {
	User map[int64]*User
}

func NewMapDataStorage() DataTxStorage {
	return &memDataMapStorage{User: make(map[int64]*User)}
}
func (storage *memDataMapStorage) GetUser(id int64) *User {
	return storage.User[id]
}
func (storage *memDataMapStorage) IterateUser(iterator func(id int64, item *User)) {
	for key, item := range storage.User {
		iterator(key, item)
	}
}
func (storage *memDataMapStorage) Apply(batch []DataLogEntity) {
	for _, tx := range batch {
		if tx.User != nil {
			switch tx.User.Action {
			case DataActionInsert, DataActionUpdate:
				storage.User[tx.User.Id] = &tx.User.Item
			case DataActionDelete:
				delete(storage.User, tx.User.Id)
			}
		}
	}
}

type User struct {
	Id       int64
	Name     string
	Email    string
	Token    string
	_project DataReader `msgp:"-"`
}

type UserLogEntity struct {
	Id     int64
	Item   User
	Action DataAction
}

type treeAdapter struct {
	tree *Tree
}

func (*treeAdapter) IterateUser(iterator func(id int64, item *User)) {
}
func (*treeAdapter) GetUser(id int64) *User {
	panic("implement me")
}
func (b *treeAdapter) Apply(batch []DataLogEntity) {
	for _, item := range batch {
		switch item.User.Action {
		case DataActionInsert:
			b.tree.Put(item.User.Id, &item.User.Item)
		}
	}
}

type btreeAdapter struct {
	tree *BTree
}

func (*btreeAdapter) GetUser(id int64) *User {
	panic("implement me")
}

func (*btreeAdapter) IterateUser(iterator func(id int64, item *User)) {
}

func (b *btreeAdapter) Apply(batch []DataLogEntity) {
	for _, item := range batch {
		switch item.User.Action {
		case DataActionInsert:
			b.tree.Put(item.User.Id, &item.User.Item)
		}
	}
}

func BenchmarkModel_defaultInsert(b *testing.B) {
	b.StopTimer()
	stor := DefaultData()

	data := make([]*User, b.N)
	for i := 0; i < b.N; i++ {
		data[i] = &User{
			Name:  "some name",
			Id:    int64(i),
			Email: "user@example.com",
			Token: "123456XXYY",
		}
	}
	tx := stor.ReadWriteLock()
	b.StartTimer()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tx.InsertUser(data[i])
	}
	tx.Commit()
}

func BenchmarkModel_treeInsert(b *testing.B) {
	b.StopTimer()
	stor := NewData(&treeAdapter{NewTree()})

	data := make([]*User, b.N)
	for i := 0; i < b.N; i++ {
		data[i] = &User{
			Name:  "some name",
			Id:    int64(i),
			Email: "user@example.com",
			Token: "123456XXYY",
		}
	}
	tx := stor.ReadWriteLock()
	b.StartTimer()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tx.InsertUser(data[i])
	}
	tx.Commit()
}

func BenchmarkModel_btreeInsert(b *testing.B) {
	b.StopTimer()
	stor := NewData(&btreeAdapter{NewBTree(128)})

	data := make([]*User, b.N)
	for i := 0; i < b.N; i++ {
		data[i] = &User{
			Name:  "some name",
			Id:    int64(i),
			Email: "user@example.com",
			Token: "123456XXYY",
		}
	}
	tx := stor.ReadWriteLock()
	b.StartTimer()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tx.InsertUser(data[i])
	}
	tx.Commit()
}
