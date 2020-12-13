# 生成Go代码

>原文：https://developers.google.com/protocol-buffers/docs/reference/go-generated

这篇文章详细描述了protocol buffer编译器为任何给定protocol定义生成Go代码的过程。对proto2和proto3生成代码之间的任何差异都会突出显示出来。在阅读本文之前，应该先阅读proto2或[proto3指南](./proto3.md)。

## Compiler 

protocol buffer编译器生成Go代码需要一个插件。
```
go install google.golang.org/protobuf/cmd/protoc-gen-go
```

将会在`$GOBIN`目录下安装一个二进制文件`protoc-gen-go`。安装目录需要在$PATH中让protocol buffer编译器可以找到插件。当指定了--go_out标志时，编译器将生成Go代码到指定的目录。编译器为每一个.profo文件生成一个源码文件，生成的源码文件的名字是对应的profo文件名字加上扩展名.pb.go。profo文件中需要包含一个`go_package`选项指定Go包的完整路径。

```
option go_package = "example.com/foo/bar";
```

生成的文件所在的输出目录的子目录取决于`go_package`选项和编译时指定的命令行参数：

* 默认情况下，生成的文件放在以Go包的导入路径命名的目录中。如，protos/foo.proto文件中有上述的go_package选项，最终生成一个example.com/foo/bar/foo.pb.go文件

* 如果给protoc设置了`--go_opt=paths=source_relative`标志，输出文件与输入文件将放在相同的目录中，如protos/foo.proto文件生成的代码放在protos/foo.pb.go。

当你像这样运行proto编译器时：

```
protoc --proto_path=src --go_out=build/gen --go_opt=paths=source_relative src/foo.proto src/bar/baz.proto
```

编译器将读取src/foo.proto和src/bar/baz.proto文件，生成两个文件build/gen/foo.pb.go和build/gen/bar/baz.pb.go

如果需要，编译器会自动创建目录build/gen/bar，但不会创建build或build/gen，它们必须已经存在。

## Packages

.proto文件应该包含一个go_package选项，指定文件的完整的Go包路径。如果没有这个选项，编译器将尝试猜测一个。之后的编译器版本将要求go_package选项必填。The Go package name of generated code will be the last path component of the go_package option.


## Messages

通过一个简单的例子分析

```
message Foo {}
```

编译器生成一个Foo的struct。*Foo实现了`proto.Message`接口。

```go
...
type Foo struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Foo) Reset()         { *m = Foo{} }
func (m *Foo) String() string { return proto.CompactTextString(m) }
func (*Foo) ProtoMessage()    {}
func (*Foo) Descriptor() ([]byte, []int) {
	return fileDescriptor_b68895c22023aa9f, []int{0}
}

func (m *Foo) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Foo.Unmarshal(m, b)
}
func (m *Foo) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Foo.Marshal(b, m, deterministic)
}
func (m *Foo) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Foo.Merge(m, src)
}
func (m *Foo) XXX_Size() int {
	return xxx_messageInfo_Foo.Size(m)
}
func (m *Foo) XXX_DiscardUnknown() {
	xxx_messageInfo_Foo.DiscardUnknown(m)
}
...
```

proto包提供对message进行操作的函数，包括二进制格式之间的转换。


`proto.Message`接口定义了一个`ProtoReflect`方法，这个方法返回一个`protoreflect.Message`代表了message的反射结果。

`optimize_for`选项不影响Go代码的生成。

**内嵌类型(Nested Types)**

```
message Foo {
  message Bar {
  }
}
```
编译器生成两个struct:Foo和Foo_Bar

```go
type Foo struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

// ......

type Foo_Bar struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}
```

## 字段(Fields)

protocol buffer编译器为message中定义的每个field生成一个struct字段。这个字段的确切性质取决于它的类型以及它是一个singular、repeated、map的字段还是oneof字段。

生成的Go字段的名称使用camel-case命名方式，即使.proto文件中的名字使用的是下划线分割的小写。转换过程如下:

1. 第一个字母大写表示导出。如果第一个字符是下划线，则删除它并在前面加上大写X。

2. 如果内部的下划线后面跟着一个小写字母，下划线将被删除，接下来的字母将被大写

按照以上规则，foo_bar_baz被转为FooBarBaz,_my_field_name_2被转为XMyFieldName_2

**Singular Scalar Fields (proto2)**

对于这两个字段定义中的任何一个:

```
optional int32 foo = 1;
required int32 foo = 1;
```

```go
type Foo struct {
	Foo                  *int32   `protobuf:"varint,1,opt,name=foo" json:"foo,omitempty"`
    Foo2                 *int32   `protobuf:"varint,2,req,name=foo2" json:"foo2,omitempty"`
}    

func (m *Foo) GetFoo() int32 {
	if m != nil && m.Foo != nil {
		return *m.Foo
	}
	return 0
}

func (m *Foo) GetFoo2() int32 {
	if m != nil && m.Foo2 != nil {
		return *m.Foo2
	}
	return 0
}
```

编译器生成一个struct,包含*int32类型的Foo字段和GetFoo()方法，方法返回一个int32类型的值或默认值


**Singular Scalar Fields (proto3)**

```
int32 foo = 1;
```

```go
type Foo struct {
	Foo                  int32    `protobuf:"varint,1,opt,name=foo,proto3" json:"foo,omitempty"`	 
}
// ......

func (m *Foo) GetFoo() int32 {
	if m != nil {
		return m.Foo
	}
	return 0
}

```

编译器将生成一个int32类型的Foo字段和一个GetFoo()方法用于获取对应的值，如果字段没有赋值则获取对应类型的默认值(数字类型是0，字符串类型是空字符)。

对于其它标量字段类型(包括bool、bytes和string)，上述的int32将会被替换为相应的Go类型。在proto文件中未设置值的字段将会被设置对应的类型的零值。


**Singular Message Fields**

定义一个message:

```
message Bar {}
```

使用用的message:

```
// proto2
message Baz {
  optional Bar foo = 1;
  // The generated code is the same result if required instead of optional.
}

// proto3
message Baz {
  Bar foo = 1;
}
```

编译器将生成如下Go代码:

```go
type Bar struct {	 
}


// proto3
type Baz struct {
	Foo                  *Bar     `protobuf:"bytes,1,opt,name=foo,proto3" json:"foo,omitempty"`
}
func (m *Baz) GetFoo() *Bar {
	if m != nil {
		return m.Foo
	}
	return nil
}
```

**Repeated Fields**

每个重复字段在Go中的结构中生成一个切片字段

```
message Baz {
  repeated Bar foo = 1;
}
```

生成如下Go代码:

```go
type Bar struct {	 
}
 
type Baz struct {
	Foo                  []*Bar   `protobuf:"bytes,1,rep,name=foo,proto3" json:"foo,omitempty"`	
}
 

func (m *Baz) GetFoo() []*Bar {
	if m != nil {
		return m.Foo
	}
	return nil
}

```

同样的，对于“repeated bytes foo = 1;”这样的定义，编译器将生成一个[][]byte类型的Foo字段；对于重复枚举定义"repeated MyEnum bar = 2;"编译器将生成[]MyNnum类型的Bar字段。

**Map Fields**

```
message Bar {}

message Baz {
  map<string, Bar> foo = 1;
}
```

生成如下代码：

```go
type Bar struct {	 
}

 
type Baz struct {
	Foo                  map[string]*Bar `protobuf:"bytes,1,rep,name=foo,proto3" json:"foo,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	 
}
 
func (m *Baz) GetFoo() map[string]*Bar {
	if m != nil {
		return m.Foo
	}
	return nil
}
```

**Oneof Fields**

对于oneof字段，pb编译器生成一个接口类型的单独字段`isMessageName_MyField`，同时也对oneof中的每个字段都生成了一个实现isMessageName_MyField接口的struct结构。

```
package account;
message Profile {
  oneof avatar {
    string image_url = 1;
    bytes image_data = 2;
  }
}
```

生成如下代码:

```go
package account

type Profile struct {
	// Types that are valid to be assigned to Avatar:
	//	*Profile_ImageUrl
	//	*Profile_ImageData
	Avatar               isProfile_Avatar `protobuf_oneof:"avatar"`
}

type isProfile_Avatar interface {
	isProfile_Avatar()
}

type Profile_ImageUrl struct {
	ImageUrl string `protobuf:"bytes,1,opt,name=image_url,json=imageUrl,proto3,oneof"`
}
func (*Profile_ImageUrl) isProfile_Avatar() {}


type Profile_ImageData struct {
	ImageData []byte `protobuf:"bytes,2,opt,name=image_data,json=imageData,proto3,oneof"`
}
func (*Profile_ImageData) isProfile_Avatar() {}

func (m *Profile) GetAvatar() isProfile_Avatar {
	if m != nil {
		return m.Avatar
	}
	return nil   //对应类型零值
}

func (m *Profile) GetImageUrl() string {
	if x, ok := m.GetAvatar().(*Profile_ImageUrl); ok {
		return x.ImageUrl
	}
	return ""
}

func (m *Profile) GetImageData() []byte {
	if x, ok := m.GetAvatar().(*Profile_ImageData); ok {
		return x.ImageData
	}
	return nil  //对应类型零值
} 
```

*Profile_ImageUrl 和 *Profile_ImageData实现了isProfile_Avatar接口，实现了接口的空方法。

赋值操作：

```go
p1 := &account.Profile{
  Avatar: &account.Profile_ImageUrl{"http://example.com/image.png"},
}

// imageData is []byte
imageData := getImageData()
p2 := &account.Profile{
  Avatar: &account.Profile_ImageData{imageData},
}
```

访问操作：

```go
switch x := m.Avatar.(type) {
case *account.Profile_ImageUrl:
        // Load profile image based on URL
        // using x.ImageUrl
case *account.Profile_ImageData:
        // Load profile image based on bytes
        // using x.ImageData
case nil:
        // The field is not set.
default:
        return fmt.Errorf("Profile.Avatar has unexpected type %T", x)
}
```


## 枚举

```
message SearchRequest {
  enum Corpus {
    UNIVERSAL = 0;
    WEB = 1;
    IMAGES = 2;
    LOCAL = 3;
    NEWS = 4;
    PRODUCTS = 5;
    VIDEO = 6;
  }
  Corpus corpus = 1;
  ...
}
```

生成代码如下:

```go
type SearchRequest_Corpus int32

const (
	SearchRequest_UNIVERSAL SearchRequest_Corpus = 0
	SearchRequest_WEB       SearchRequest_Corpus = 1
	SearchRequest_IMAGES    SearchRequest_Corpus = 2
	SearchRequest_LOCAL     SearchRequest_Corpus = 3
	SearchRequest_NEWS      SearchRequest_Corpus = 4
	SearchRequest_PRODUCTS  SearchRequest_Corpus = 5
	SearchRequest_VIDEO     SearchRequest_Corpus = 6
)

type SearchRequest struct {
	Corpus               SearchRequest_Corpus `protobuf:"varint,1,opt,name=corpus,proto3,enum=SearchRequest_Corpus" json:"corpus,omitempty"`

}

func (m *SearchRequest) GetCorpus() SearchRequest_Corpus {
	if m != nil {
		return m.Corpus
	}
	return SearchRequest_UNIVERSAL
}
```

对于message内部的枚举生成的枚举名字是以message的名字开头的，对于包级别的枚举其名字就是定义的名字。

proto允许多个枚举符号具有相同的字段编号，具有相同编号的符号是同义词。它们在Go中以完全相同的方式表示，即多个名称对应于同一个数值。

```
enum Foo {
   option allow_alias=true;
   DEFAULT_BAR = 0;
   BAR_BELLS = 1;
   BAR_B_CUE = 1;
}
```

生成代码:

```go
type Foo int32

const (
	Foo_DEFAULT_BAR Foo = 0
	Foo_BAR_BELLS   Foo = 1
	Foo_BAR_B_CUE   Foo = 1
)
```

## Extensions (proto2)

略

## Services

默认情况下，进行Go代码生成时不会生成service。如果启用gRPC插件，则会生成支持gRPC的service。

