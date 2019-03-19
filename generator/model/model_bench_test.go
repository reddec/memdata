package model

import "testing"

type UserStorage interface {
	PutUser(id int64, item *User)
	GetUser(id int64) *User
	DeleteUser(id int64)
	IterateUser(iterator func(id int64, item *User))
}

type Data struct {
	sequenceUserId int64
	indexUserById  UserStorage
}

func NewData(storageUserById UserStorage) *Data {
	// restore auto-sequence for User.Id
	var maxIdOfUser int64
	storageUserById.IterateUser(func(id int64, item *User) {
		if id > maxIdOfUser {
			maxIdOfUser = id
		}
	})
	return &Data{indexUserById: storageUserById, sequenceUserId: maxIdOfUser}
}
func DefaultData() *Data {
	return NewData(NewMapUserStorage())
}
func (project *Data) User(id int64) *User {
	return project.indexUserById.GetUser(id)
}
func (project *Data) NextUserId() int64 {
	project.sequenceUserId++
	return project.sequenceUserId
}
func (project *Data) InsertUser(item *User) *User {
	item.Id = project.NextUserId()
	item._project = project
	project.indexUserById.PutUser(item.Id, item)
	return item
}
func (project *Data) RemoveUser(id int64) {
	project.indexUserById.DeleteUser(id)
}

type mapUserStorage struct {
	data map[int64]*User
}

func NewMapUserStorage() UserStorage {
	return &mapUserStorage{data: make(map[int64]*User)}
}
func (storage *mapUserStorage) PutUser(id int64, item *User) {
	storage.data[id] = item
}
func (storage *mapUserStorage) GetUser(id int64) *User {
	return storage.data[id]
}
func (storage *mapUserStorage) DeleteUser(id int64) {
	delete(storage.data, id)
}
func (storage *mapUserStorage) IterateUser(iterator func(id int64, item *User)) {
	for key, item := range storage.data {
		iterator(key, item)
	}
}

type User struct {
	Token    string
	Id       int64
	Name     string
	Email    string
	_project *Data `msgp:"-"`
}

func (model *User) Data() *Data {
	return model._project
}

type arrayStore struct {
	data []*User
}

func (a *arrayStore) PutUser(id int64, item *User) {
	a.data = append(a.data, item)
}

func (*arrayStore) GetUser(id int64) *User {
	panic("implement me")
}

func (*arrayStore) DeleteUser(id int64) {
	panic("implement me")
}

func (*arrayStore) IterateUser(iterator func(id int64, item *User)) {}

type treeAdapter struct {
	tree *Tree
}

func (ta *treeAdapter) PutUser(id int64, item *User) {
	ta.tree.Put(id, item)
}

func (*treeAdapter) GetUser(id int64) *User {
	panic("implement me")
}

func (*treeAdapter) DeleteUser(id int64) {
	panic("implement me")
}

func (*treeAdapter) IterateUser(iterator func(id int64, item *User)) {

}

type btreeAdapter struct {
	tree *BTree
}

func (ta *btreeAdapter) PutUser(id int64, item *User) {
	ta.tree.Put(id, item)
}

func (*btreeAdapter) GetUser(id int64) *User {
	panic("implement me")
}

func (*btreeAdapter) DeleteUser(id int64) {
	panic("implement me")
}

func (*btreeAdapter) IterateUser(iterator func(id int64, item *User)) {

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
	b.StartTimer()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stor.InsertUser(data[i])
	}
}

func BenchmarkModel_arrayInsert(b *testing.B) {
	b.StopTimer()
	stor := NewData(&arrayStore{})

	data := make([]*User, b.N)
	for i := 0; i < b.N; i++ {
		data[i] = &User{
			Name:  "some name",
			Id:    int64(i),
			Email: "user@example.com",
			Token: "123456XXYY",
		}
	}
	b.StartTimer()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stor.InsertUser(data[i])
	}
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
	b.StartTimer()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stor.InsertUser(data[i])
	}
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
	b.StartTimer()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stor.InsertUser(data[i])
	}
}
