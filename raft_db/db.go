package raft_db

import (
	"encoding/json"
	"errors"
)

type Change struct {
	TypeModification int    `json:"type_modification"`
	Key              string `json:"key"`
	Value            string `json:"value"`
}

const (
	Create = iota
	Update
	Delete
)

func NewChange(typeModification int, key, value string) *Change {
	return &Change{
		TypeModification: typeModification,
		Key:              key,
		Value:            value,
	}
}

type LocalDataBase struct {
	DBVariables map[string]string `json:"db_variables"`
	Changelog   []*Change         `json:"changelog"`
	iterator 	int
}

func NewLocalDataBase() *LocalDataBase {
	return &LocalDataBase{
		DBVariables: make(map[string]string),
		Changelog:   make([]*Change, 0),
		iterator: 	 0,
	}
}

func (db *LocalDataBase) Create(key, value string) {
	db.DBVariables[key] = value
	change := NewChange(Create, key, value)
	db.Changelog = append(db.Changelog, change)
}

func (db *LocalDataBase) Read(key string) (string, error) {
	value, exists := db.DBVariables[key]
	if !exists {
		return "", errors.New("key not found")
	}
	return value, nil
}

func (db *LocalDataBase) Update(key, value string) error {
	_, exists := db.DBVariables[key]
	if !exists {
		return errors.New("key not found")
	}
	db.DBVariables[key] = value
	change := NewChange(Update, key, value)
	db.Changelog = append(db.Changelog, change)
	return nil
}

func (db *LocalDataBase) Delete(key string) error {
	_, exists := db.DBVariables[key]
	if !exists {
		return errors.New("key not found")
	}
	delete(db.DBVariables, key)
	change := NewChange(Delete, key, "")
	db.Changelog = append(db.Changelog, change)
	return nil
}

func (db *LocalDataBase) Serialize(fromPosition int) ([]byte, error) {
	if fromPosition < 0 || fromPosition >= len(db.Changelog) {
		return nil, errors.New("invalid changelog position")
	}

	serializedDB := &LocalDataBase{
		Changelog:   db.Changelog[fromPosition:],
	}

	return json.Marshal(serializedDB)
}

func Deserialize(data []byte) (*LocalDataBase, error) {
	var db LocalDataBase
	err := json.Unmarshal(data, &db)
	if err != nil {
		return nil, err
	}
	return &db, nil
}

func (db *LocalDataBase) AppendNewChanges(another *LocalDataBase) {
	for _, change := range another.Changelog {
		switch change.TypeModification {
		case Create:
			db.DBVariables[change.Key] = change.Value
		case Update:
			db.DBVariables[change.Key] = change.Value
		case Delete:
			delete(db.DBVariables, change.Key)
		}
		db.Changelog = append(db.Changelog, change)
	}
}

func (db *LocalDataBase) GetNextPartOfData() ([]*Change, error) {
	changelog_to_send := db.Changelog[db.iterator:]
	db.iterator = len(db.Changelog)
	return changelog_to_send, nil
}

func (db *LocalDataBase) GetChangelogFromPosition(position int) ([]*Change, error) {
	changelog_to_send := db.Changelog[position:]
	return changelog_to_send, nil
}

func (db *LocalDataBase) Size() int {
	return len(db.Changelog)
}

func DeserializeChanges(data []byte) ([]*Change, error) {
	var changes []*Change
	err := json.Unmarshal(data, &changes)
	if err != nil {
		return nil, err
	}
	return changes, nil
}
