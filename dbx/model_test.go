package dbx

import (
	"database/sql/driver"
	"encoding/json"
	"reflect"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/ml444/gkit/dbx/pagination"
	"github.com/ml444/gkit/pkg/cryptox"
)

/*
var dbUser = dbx.NewT[user.ModelUser](dbx.DB, dbx.WithNotFoundErrCode(user.ErrNotFoundUser),
	dbx.WithTableCipher(config.GetDefaultAESCipher()),
	dbx.WithNeedDecrypt(true), // 业务层透明，无需业务层关心解密逻辑
	dbx.WithSpecifyFieldCipherMap(config.GetTableFieldsCipher(new(user.ModelUser).TableName())),
)
*/

func Benchmark_Reflect_SetStructField(b *testing.B) {
	s := &testModel{
		ID:   1,
		Name: "test",
	}
	mVV := reflect.ValueOf(s)
	mV := mVV.Elem().FieldByName("Name")
	for i := 0; i < b.N; i++ {
		switch mV.Interface().(type) {
		case string:
			if mV.CanSet() {
				mV.SetString("test2")
			} else {
				panic("can't set")
			}
		}
	}
}

func Benchmark_Reflect_SetMapField(b *testing.B) {
	s := map[string]any{
		"ID":   1,
		"Name": "test",
	}
	mVV := reflect.ValueOf(s)
	//  mV := mVV.MapIndex(reflect.ValueOf("Name"))
	//  b.Log("====> ", mV.Interface())
	for i := 0; i < b.N; i++ {
		mVV.SetMapIndex(reflect.ValueOf("Name"), reflect.ValueOf("test2"))
		//if mVV.CanSet() {
		//	mVV.SetMapIndex(mV, reflect.ValueOf("test2"))
		//} else {
		//	panic("can't set")
		//}
		//if mVV.MapIndex(reflect.ValueOf("Name")).Interface() != "test2" {
		//	b.Errorf("expected 'test2', got '%v'", mVV.MapIndex(reflect.ValueOf("Name")).Interface())
		//}
	}
}

type testModel struct {
	ID        uint64       `json:"id"`
	CreatedAt uint32       `gorm:"comment:创建时间" json:"created_at"`
	UpdatedAt uint32       `gorm:"comment:更新时间" json:"updated_at"`
	DeletedAt uint32       `gorm:"comment:删除时间" json:"deleted_at"`
	Name      string       `json:"name"`
	Addresses []string     `json:"addresses"`
	ParentID  uint64       `json:"parent_id"`
	Parent    *testModel   `json:"parent"`
	Children  []*testModel `json:"children"`
}

func (t *testModel) TableName() string {
	return "test_model"
}

type testOrmModel struct {
	ID        uint64              `json:"id" gorm:"primaryKey"`
	CreatedAt uint32              `gorm:"comment:创建时间" json:"created_at"`
	UpdatedAt uint32              `gorm:"comment:更新时间" json:"updated_at"`
	DeletedAt uint32              `gorm:"comment:删除时间" json:"deleted_at"`
	Name      string              `json:"name" gorm:"column:name;type:varchar(255);not null;encrypt:true"`
	Addresses testModel_Addresses `json:"addresses" gorm:"column:addresses;type:json"`
	ParentID  uint64              `json:"parent_id"`
	Parent    *testOrmModel       `json:"parent" gorm:"foreignKey:ParentID"`
	Children  testModel_Children  `json:"children" gorm:"type:json"`
}

func (t *testOrmModel) TableName() string {
	return "test_model"
}

type testModel_Addresses []string

func (t *testModel_Addresses) Scan(value interface{}) error {
	return json.Unmarshal(value.([]byte), t)
}

func (t testModel_Addresses) Value() (driver.Value, error) {
	return json.Marshal(t)
}

type testModel_Children []*testModel

func (t *testModel_Children) Scan(value interface{}) error {
	return json.Unmarshal(value.([]byte), t)
}

func (t testModel_Children) Value() (driver.Value, error) {
	return json.Marshal(t)
}

func (t *testModel) ToORM() ITModel {
	var parent *testOrmModel
	if t.Parent != nil {
		parent = t.Parent.ToORM().(*testOrmModel)
	}

	return &testOrmModel{
		ID:        t.ID,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
		DeletedAt: t.DeletedAt,
		Name:      t.Name,
		Addresses: testModel_Addresses(t.Addresses),
		ParentID:  t.ParentID,
		Parent:    parent,
		Children:  testModel_Children(t.Children),
	}
}

func (x *testOrmModel) ToSource() IModel {
	return &testModel{
		ID:        x.ID,
		CreatedAt: x.CreatedAt,
		UpdatedAt: x.UpdatedAt,
		DeletedAt: x.DeletedAt,
		Name:      x.Name,
		Addresses: []string(x.Addresses),
	}
}

func (x *testOrmModel) ForceTModel() bool {
	return true
}

var tx *gorm.DB

func testGetDB() *gorm.DB {
	if tx == nil {
		db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
		if err != nil {
			panic(err)
		}
		tx = db
		db.AutoMigrate(&testOrmModel{})
	}
	return tx
}

func testGetCipher() ICipher {
	cryptor, err := cryptox.NewAES(
		[]byte("1234567890123456"),
		cryptox.AESOptWithFixedNonce([]byte("123456789098")),
	)
	if err != nil {
		panic(err)
	}
	return cryptor
}

func TestNewT(t *testing.T) {
	type args struct {
		opts []TOption
	}
	tests := []struct {
		name string
		args args
		want *T
	}{
		{
			name: "ok",
			args: args{
				opts: []TOption{
					SetCreateBatchSize(50),
					SetTableCipher(testGetCipher()),
					SetDisableDecrypt(),
					SetSpecifyFieldCipherMap(map[string]FieldCipher{}),
					SetNotFoundErrCode(40000),
					SetPrimaryKey("pk"),
					SetGenerateIDFunc(func() uint64 { return 0 }),
				},
			},
			want: &T{
				getDB:           testGetDB,
				model:           &testModel{},
				ormModel:        &testOrmModel{},
				forceTModel:     true,
				NotFoundErrCode: 40000,
				BatchCreateSize: 50,
				PrimaryKey:      "pk",
				EncryptFieldMap: map[string]ICipher{"name": testGetCipher(), "Name": testGetCipher()},
				NeedEncrypt:     true,
				DisableDecrypt:  true,
				IdGenerator:     func() uint64 { return 0 },
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewT[testModel](testGetDB, tt.args.opts...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewT() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestT_CheckAndEncrypt(t *testing.T) {
	type args struct {
		m        any
		kind     cipherKind
		isCreate bool
	}

	tests := []struct {
		name    string
		opts    []TOption
		args    args
		want    any
		wantErr bool
	}{
		{
			name:    "ok_struct",
			opts:    []TOption{SetTableCipher(testGetCipher())},
			args:    args{&testModel{ID: 123, Name: "test"}, cipherKindEncrypt, false},
			want:    &testModel{ID: 123, Name: "MTIzNDU2Nzg5MDk4hRwcyLRJ5UF8B0knCxfpcMYxjrY"},
			wantErr: false,
		},
		{
			name:    "ok_struct_specify_field",
			opts:    []TOption{SetTableCipher(testGetCipher()), SetSpecifyFieldCipherMap(map[string]FieldCipher{"addresses": {"Addresses", testGetCipher()}})},
			args:    args{&testModel{ID: 123, Name: "test", Addresses: []string{"foo", "bar"}}, cipherKindEncrypt, false},
			want:    &testModel{ID: 123, Name: "MTIzNDU2Nzg5MDk4hRwcyLRJ5UF8B0knCxfpcMYxjrY", Addresses: []string{"MTIzNDU2Nzg5MDk4lxYASANRd9OSafwxd9W3GsU9EQ", "MTIzNDU2Nzg5MDk4kxgdFWKbhWqdp7u7BHN/IWvokg"}},
			wantErr: false,
		},
		{
			name:    "ok_struct_empty",
			opts:    []TOption{SetTableCipher(testGetCipher())},
			args:    args{&testModel{ID: 123, Name: "", Addresses: nil}, cipherKindEncrypt, false},
			want:    &testModel{ID: 123, Name: "", Addresses: nil},
			wantErr: false,
		},
		{
			name:    "ok_map",
			opts:    []TOption{SetTableCipher(testGetCipher())},
			args:    args{map[string]any{"id": 123, "name": "test"}, cipherKindEncrypt, false},
			want:    map[string]any{"id": 123, "name": "MTIzNDU2Nzg5MDk4hRwcyLRJ5UF8B0knCxfpcMYxjrY"},
			wantErr: false,
		},
		{
			name:    "ok_map_specify_field",
			opts:    []TOption{SetTableCipher(testGetCipher()), SetSpecifyFieldCipherMap(map[string]FieldCipher{"addresses": {"", testGetCipher()}})},
			args:    args{map[string]any{"id": 123, "name": "test", "addresses": []string{"foo", "bar"}}, cipherKindEncrypt, false},
			want:    map[string]any{"id": 123, "name": "MTIzNDU2Nzg5MDk4hRwcyLRJ5UF8B0knCxfpcMYxjrY", "addresses": []string{"MTIzNDU2Nzg5MDk4lxYASANRd9OSafwxd9W3GsU9EQ", "MTIzNDU2Nzg5MDk4kxgdFWKbhWqdp7u7BHN/IWvokg"}},
			wantErr: false,
		},
		{
			name:    "ok_map_empty",
			opts:    []TOption{SetTableCipher(testGetCipher()), SetSpecifyFieldCipherMap(map[string]FieldCipher{"addresses": {"", testGetCipher()}})},
			args:    args{map[string]any{"name": "", "addresses": []string{}}, cipherKindEncrypt, false},
			want:    map[string]any{"name": "", "addresses": []string{}},
			wantErr: false,
		},
		{
			name:    "ok_empty_map",
			opts:    []TOption{SetTableCipher(testGetCipher()), SetSpecifyFieldCipherMap(map[string]FieldCipher{"addresses": {"", testGetCipher()}})},
			args:    args{map[string]any{}, cipherKindEncrypt, false},
			want:    map[string]any{},
			wantErr: false,
		},
		{
			name:    "ok_map_decrypt",
			opts:    []TOption{SetTableCipher(testGetCipher())},
			args:    args{map[string]string{"name": "MTIzNDU2Nzg5MDk4hRwcyLRJ5UF8B0knCxfpcMYxjrY"}, cipherKindDecrypt, false},
			want:    map[string]string{"name": "test"},
			wantErr: false,
		},
		{
			name:    "ok_map_specify_field_decrypt",
			opts:    []TOption{SetTableCipher(testGetCipher()), SetSpecifyFieldCipherMap(map[string]FieldCipher{"addresses": {"", testGetCipher()}})},
			args:    args{map[string]any{"id": 123, "name": "MTIzNDU2Nzg5MDk4hRwcyLRJ5UF8B0knCxfpcMYxjrY", "addresses": []string{"MTIzNDU2Nzg5MDk4lxYASANRd9OSafwxd9W3GsU9EQ", "MTIzNDU2Nzg5MDk4kxgdFWKbhWqdp7u7BHN/IWvokg"}}, cipherKindDecrypt, false},
			want:    map[string]any{"id": 123, "name": "test", "addresses": []string{"foo", "bar"}},
			wantErr: false,
		},
		{
			name:    "ok_slice",
			opts:    []TOption{SetTableCipher(testGetCipher())},
			args:    args{m: []*testModel{{ID: 1, Name: "test"}, {ID: 2, Name: "test2"}}},
			want:    []*testModel{{ID: 1, Name: "MTIzNDU2Nzg5MDk4hRwcyLRJ5UF8B0knCxfpcMYxjrY"}, {ID: 2, Name: "MTIzNDU2Nzg5MDk4hRwcyC8+/YT06otATC5KMGl1cnIj"}},
			wantErr: false,
		},
		{
			name:    "ok_slice_ptr",
			opts:    []TOption{SetTableCipher(testGetCipher())},
			args:    args{m: &[]*testModel{{ID: 1, Name: "test"}, {ID: 2, Name: "test2"}}},
			want:    &[]*testModel{{ID: 1, Name: "MTIzNDU2Nzg5MDk4hRwcyLRJ5UF8B0knCxfpcMYxjrY"}, {ID: 2, Name: "MTIzNDU2Nzg5MDk4hRwcyC8+/YT06otATC5KMGl1cnIj"}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			x := NewT[testModel](testGetDB, tt.opts...)
			if err := x.CheckAndCrypto(tt.args.m, tt.args.kind, tt.args.isCreate); (err != nil) != tt.wantErr {
				t.Errorf("CheckAndCrypto() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(tt.args.m, tt.want) {
				if mm, ok := tt.args.m.([]*testModel); ok {
					for _, v := range mm {
						t.Errorf("===> value: %#v", v)
					}
				}
				t.Errorf("CheckAndCrypto() got = %#v, want %#v", tt.args.m, tt.want)
			}
		})
	}
}

func TestT_Create(t *testing.T) {
	tests := []struct {
		name    string
		opts    []TOption
		m       any
		wantErr bool
	}{
		{name: "Create", m: &testModel{Name: "test", Addresses: []string{"abc", "efg"}}, wantErr: false},
		{name: "Create_1", m: &testModel{Name: "test2", Addresses: []string{"abc", "efg"}, ParentID: 1, Children: []*testModel{{Name: "foo"}, {Name: "bar"}}}, wantErr: false},
		{name: "Create_map", m: map[string]any{"name": "test3", "parent_id": 1}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			x := NewT[testModel](testGetDB, tt.opts...)
			if err := x.Create(tt.m); (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}
			if mV, ok := tt.m.(*testModel); ok {
				if mV.ID == 0 {
					t.Errorf("Create() failed: %#v", tt.m)
				} else {
					var m testModel
					err := x.GetOne(mV.ID, &m)
					if err != nil {
						t.Error(err.Error())
					} else {
						t.Logf("Success get one: %+v", &m)
					}

				}
			} else if mVV, okk := tt.m.(map[string]any); okk {
				t.Log(mVV)
			}
		})
	}
}

func TestT_BatchCreate(t *testing.T) {
	type args struct {
		list any
	}
	tests := []struct {
		name    string
		opts    []TOption
		list    any
		wantErr bool
	}{
		{
			name:    "ok",
			opts:    []TOption{},
			list:    []*testModel{{Name: "test1"}, {Name: "test2"}, {Name: "test3"}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			x := NewT[testModel](testGetDB, tt.opts...)
			if err := x.BatchCreate(&tt.list); (err != nil) != tt.wantErr {
				t.Errorf("BatchCreate() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				t.Logf("BatchCreate success: %+v", tt.list)
			}
			var idList []uint64
			for _, m := range tt.list.([]*testModel) {
				idList = append(idList, m.ID)
			}
			cnt, err := x.Count("id", idList)
			if err != nil {
				t.Fatal(err)
			}
			if cnt != int64(len(tt.list.([]*testModel))) {
				t.Errorf("===> insert affect row len: %d", cnt)
			}
		})
	}
}

func TestT_Count(t *testing.T) {
	tests := []struct {
		name    string
		want    int64
		wantErr bool
	}{
		{
			name:    "",
			want:    3,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			x := NewT[testModel](testGetDB)
			x.BatchCreate([]*testModel{{Name: "test1"}, {Name: "test2"}, {Name: "test3"}})
			got, err := x.Count()
			if (err != nil) != tt.wantErr {
				t.Errorf("Count() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got == 0 {
				t.Errorf("Count() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestT_GetOne_and_Delete_and_Exist(t *testing.T) {
	x := NewT[testModel](testGetDB, SetNotFoundErrCode(1000))
	m := testModel{Name: "testGet", Addresses: []string{"foo", "bar"}}
	err := x.Create(&m)
	if err != nil {
		t.Fatal(err)
	}
	var m1 testModel
	err = x.GetOne(m.ID, &m1)
	if err != nil {
		t.Fatal(err)
	}
	if m1.ID == 0 || m1.Name != m.Name || !reflect.DeepEqual(m.Addresses, m1.Addresses) {
		t.Fatalf("GetOne is err, got: %+v", m1)
	}
	var m2 testModel
	err = x.GetOneByWhere(&m2, "name", m.Name)
	if err != nil {
		t.Fatal(err)
	}
	if m2.ID == 0 || m2.Name != m.Name || !reflect.DeepEqual(m.Addresses, m2.Addresses) {
		t.Fatalf("GetOne is err, got: %+v", m2)
	}
	var m3 testModel
	err = x.GetOneByWhere(&m3, map[string]string{"name": m.Name})
	if err != nil {
		t.Fatal(err)
	}
	if m3.ID == 0 || m3.Name != "testGet" || !reflect.DeepEqual(m.Addresses, m3.Addresses) {
		t.Fatalf("GetOne is err, got: %+v", m2)
	}
	if err := x.DeleteByPk(m.ID); err != nil {
		t.Errorf("DeleteByPk() error = %v", err)
	} else {
		exist, err := x.ExistByWhere("id", m.ID)
		if err != nil {
			t.Fatal(err)
		}
		if exist {
			t.Fatal("this data has been deleted, shouldn't exist")
		}
	}

	mDel := testModel{Name: "testDeleteByWhere", Addresses: []string{"foo", "bar"}}
	err = x.Create(&mDel)
	if err != nil {
		t.Fatal(err)
	}
	err = x.DeleteByWhere("name", mDel.Name)
	if err != nil {
		t.Fatal(err)
	}
	err = x.DeleteByWhere(map[string]string{"name": mDel.Name})
	if err != nil {
		t.Fatal(err)
	}
}

func TestT_ListAll(t *testing.T) {
	x := NewT[testModel](testGetDB)
	mList := []*testModel{
		{Name: "testListAll1", Addresses: []string{"add1", "addr11"}},
		{Name: "testListAll2", Addresses: []string{"add2", "addr22"}, ParentID: 1},
		{Name: "testListAll3", Addresses: []string{"add3", "addr33"}},
	}
	err := x.BatchCreate(&mList)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name      string
		opts      any
		wantTotal int
		wantErr   bool
	}{
		{
			name:    "invaild_opts",
			opts:    &testModel{Name: "test"},
			wantErr: true,
		},
		{
			name:      "map_query",
			opts:      map[string]string{"name": "testListAll1"},
			wantTotal: 1,
			wantErr:   false,
		},
		{
			name: "QueryOpt",
			opts: &QueryOpts{
				Where: map[string]any{"parent_id": 0},
				Like:  map[string]string{"name": "testListAll"},
			},
			wantTotal: 2,
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var list []*testModel
			if err := x.ListAll(tt.opts, &list); (err != nil) != tt.wantErr {
				t.Errorf("ListAll() error = %v, wantErr %v", err, tt.wantErr)
			}
			if len(list) != tt.wantTotal {
				t.Fatalf("got: %+v", list)
			}
			if len(list) > 0 {
				m := list[0]
				if len(m.Addresses) == 0 {
					t.Fatal("fetch data err")
				}
			}
		})
	}
}

func TestT_ListWithPagination(t *testing.T) {
	type args struct {
		paginate *pagination.Pagination
		opts     interface{}
		listPtr  interface{}
	}
	tests := []struct {
		name    string
		opts    []TOption
		args    args
		want    *pagination.Pagination
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			x := NewT[testModel](testGetDB, tt.opts...)
			got, err := x.ListWithPagination(tt.args.paginate, tt.args.opts, tt.args.listPtr)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListWithPagination() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ListWithPagination() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestT_Scope(t *testing.T) {
	tests := []struct {
		name string
		opts []TOption
		want *Scope
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			x := NewT[testModel](testGetDB, tt.opts...)
			if got := x.Scope(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Scope() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestT_Update(t *testing.T) {
	type args struct {
		m        interface{}
		whereMap map[string]interface{}
	}
	tests := []struct {
		name     string
		opts     []TOption
		args     args
		wantRows int64
		wantErr  bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			x := NewT[testModel](testGetDB, tt.opts...)
			gotRows, err := x.Update(tt.args.m, tt.args.whereMap)
			if (err != nil) != tt.wantErr {
				t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotRows != tt.wantRows {
				t.Errorf("Update() gotRows = %v, want %v", gotRows, tt.wantRows)
			}
		})
	}
}

func TestT_validateListAndGetModel(t *testing.T) {
	type args struct {
		listPtr interface{}
	}
	tests := []struct {
		name    string
		opts    []TOption
		args    args
		want    interface{}
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			x := NewT[testModel](testGetDB, tt.opts...)
			got, err := x.validateListAndGetModel(tt.args.listPtr)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateListAndGetModel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("validateListAndGetModel() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_camelToSnake(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := camelToSnake(tt.args.s); got != tt.want {
				t.Errorf("camelToSnake() = %v, want %v", got, tt.want)
			}
		})
	}
}
