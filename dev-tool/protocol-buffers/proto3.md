
# proto3语法指南

>原文:https://developers.google.com/protocol-buffers/docs/proto3


## 定义消息类型

先看一个简单的例子。假设您想要定义一个搜索请求消息格式，其中每个搜索请求都有一个查询字符串、您感兴趣的特定结果页以及每页的结果数量。下面是对应的`.proto`文件:

```
syntax = "proto3";

message SearchRequest {
  string query = 1;
  int32 page_number = 2;
  int32 result_per_page = 3;
}
```

* 第一行用来声明使用的是proto3语法，如果没有这一行，pb编译器将认为使用的prot2语法。它是是文件的第一个非空、非注释行。

* SearchRequest消息定义中指定了三个字段,每个字段都有一个名称和类型

**字段类型**

上面例子中，所有字段都是变量类型:两个正数和一个字符串。你可以指定为字段指定复杂类型，包括枚举或其它类型

**字段编号**

如你所见，消息中定义的每个字段都有一个唯一的编号。这些字段的编号用于在二进制格式的消息中标记字段，一旦定义的类型被使用就不应该在改变对应的编号。编号从1到15的字段使用一个字节对编号和类型进行编码。编号在16到2047之间的使用两个字节。因此应该将1到15预留给经常使用的字段。请记住为将来可能添加的频繁出现的字段留出一些空间。

字段编号最小是1，最大是2^29-1(536,870,911)。19000到19999之间的编号是pb本身保留的不能使用，如果使用了保留的编号，pb编译器会提示。

**字段规则**

消息中的字段可以具有以下修饰

* singular：字段可以最多出现一次(即可选),这是proto3语法的默认字段规则。

* repeated：字段可以重复任意次数(包括零)。

proto3中，重复字段的标量数字字段默认进行压缩编码

**更多信息类型**

多个消息类型可以定义在同一个.proto文件中，如果要定义多个相关消息这种方式很有用。如:

```
message SearchRequest {
  string query = 1;
  int32 page_number = 2;
  int32 result_per_page = 3;
}

message SearchResponse {
 ...
}
```

**注释**

在.proto文件中添加注释，可以使用`//`和`/*...*/`方式

```
/* SearchRequest represents a search query, with pagination options to
 * indicate which results to include in the response. */

message SearchRequest {
  string query = 1;
  int32 page_number = 2;  // Which page number do we want?
  int32 result_per_page = 3;  // Number of results to return per page.
}
```

**保留字段**

如果你通过删除字段或注释字段来修改消息类型，那么将来其它用户在更新时会重用该字段编号。如果后边在使用到这个.proto文件会导致严重的问题，包括数据损坏、隐私漏洞等。通过将已删的字段的编号保留的方式可以避免发生这样的情况。如果使用了保留字段，编译器会提示。

```
message Foo {
  reserved 2, 15, 9 to 11;
  reserved "foo", "bar";
}
```

需要注意不能在reserved中混合使用字段名和字段编号。

**根据.proto生成了什么**

当使用protocol buffer编译器编译.proto文件时，编译器根据选择的语言和文件中定义的消息类型生成代码，包括字段值的get/set、将消息序列化到输出流和从输入流解析消息。

* C++,编译器生成.h和.cc文件，为定义的消息生成一个class

* Java,编译器生成.java文件，为定义的消息生成一个class,同时生成用于创建类实例的Builder

* Go,编译器生成.pb.go文件，为定义的消息生成一个对应的类型


## 标量值类型

标量消息字段可以具有以下类型之一，下表列出了.proto文件中的类型在生成的对应语言中的类型

| .proto类型 | 说明 | Java类型   | Go类型  | C#类型     | ...  |
| ---------- | ---- | ---------- | ------- | ---------- | ---- |
| double     |      | double     | float64 | double     |      |
| float      |      | float      | float32 | float      |      |
| int32      |      | int        | int32   | int        |      |
| int64      |      | long       | int64   | long       |      |
| uint32     |      | int        | uint32  | uint       |      |
| uint64     |      | long       | uint64  | ulong      |      |
| sint32     |      | int        | int32   | int        |      |
| sint64     |      | long       | int64   | long       |      |
| fixed32    |      | int        | uint32  | uint       |      |
| fixed64    |      | long       | uint64  | ulong      |      |
| sfixed32   |      | int        | int32   | int        |      |
| sfixed64   |      | long       | int64   | long       |      |
| bool       |      | boolean    | bool    | bool       |      |
| string     |      | String     | string  | string     |      |
| bytes      |      | ByteString | []byte  | ByteString |      |


## 默认值

不同类型的默认值:

* string：默认空字符串

* byte：默认空字节

* bool：默认false

* 数字类型：默认0

* 枚举：默认定义的第一个枚举值，编码必须是0

* 自定义消息类型：取决与语言


## 枚举

```
message SearchRequest {
  string query = 1;
  int32 page_number = 2;
  int32 result_per_page = 3;
  enum Corpus {
    UNIVERSAL = 0;
    WEB = 1;
    IMAGES = 2;
    LOCAL = 3;
    NEWS = 4;
    PRODUCTS = 5;
    VIDEO = 6;
  }
  Corpus corpus = 4;
}
```

枚举类型第一个元素的值必须是0，这是因为：

* 作为默认值

* 零值需要作为定义一个元素是为了兼容proto2语义

可以通过将相同的值分配给不同的枚举常量来定义别名。需要设置`allow_alias`为true才能使用这项功能，否则会报错。

```
message MyMessage1 {
  enum EnumAllowingAlias {
    option allow_alias = true;
    UNKNOWN = 0;
    STARTED = 1;
    RUNNING = 1;
  }
}
message MyMessage2 {
  enum EnumNotAllowingAlias {
    UNKNOWN = 0;
    STARTED = 1;
    // RUNNING = 1;  // Uncommenting this line will cause a compile error inside Google and a warning message outside.
  }
}
```

枚举器常量必须在32位整数范围内，由于编码方式负值效率低下因此不建议使用负值。枚举可以定义在消息类型内部可以定义在外部，外部的定义的枚举可以被其它消息复用。还可以将一条消息中声明的枚举类型用作另一条消息中字段的类型，使用语法如下：`_MessageType_._EnumType_`


**保留值**

编号一旦被使用，如果后边被修改，可能会出现异常。通过保留这些编号，pb编译器在编译时会检查是否使用了保留的编号从而给出提示。可以使用`max`关键字指定保留的数值范围增加到最大可能值。

```
enum Foo {
  reserved 2, 15, 9 to 11, 40 to max;
  reserved "FOO", "BAR";
}
```

> 请注意，不能在同一个保留语句中混合字段名和字段编号。


## 使用其它消息类型

你可以使用其它消息类型作为字段类型。

```
message SearchResponse {
  repeated Result results = 1;
}

message Result {
  string url = 1;
  string title = 2;
  repeated string snippets = 3;
}
```

**导入定义**

在上面例子中，Result和SearchResponse消息类型定义在相同的文件中。如果要用的消息类型定义在了其它`.proto`文件，可以通过导入的方式来使用其它文件中的消息类型。

```
import "myproject/other_protos.proto";
```

默认情况下，只能使用直接导入的.proto文件中的定义。然而，有时你可能将.proto文件移动到一个新的位置,这样的话还需要更新使用到这个文件的地方。现在可以在旧位置放置一个假的.proto文件将所有导入转发到新位置。

```
// new.proto
// All definitions are moved here
```

```
// old.proto
// This is the proto that all clients are importing.
import public "new.proto";
import "other.proto";
```
```
// client.proto
import "old.proto";
// You use definitions from old.proto and new.proto, but not other.proto
```

PB编译器在通过`-I/--proto_path`执行的目录下查找导入的文件，如果没有指定目录则默认在编译器被调用的位置处查找。


**使用proto2消息类型**

可以导入proto2消息类型并在你的proto3消息中使用它们，反之亦然。但是，proto2中的枚举不能在proto3语法中直接使用(proto3的枚举可以在proto2中直接使用)


## 嵌套类型

可以在消息类型中定义和使用其它消息类型，即将一个类型内嵌在其它类型内:

```
message SearchResponse {
  message Result {
    string url = 1;
    string title = 2;
    repeated string snippets = 3;
  }
  repeated Result results = 1;
}
```

如果你先重用内嵌的消息类型，需要使用这样的格式:`_Parent_._Type_`

```
message SomeOtherMessage {
  SearchResponse.Result result = 1;
}
```

内嵌的深度，可以根据你的需要进行

```
message Outer {                  // Level 0
  message MiddleAA {  // Level 1
    message Inner {   // Level 2
      int64 ival = 1;
      bool  booly = 2;
    }
  }
  message MiddleBB {  // Level 1
    message Inner {   // Level 2
      int32 ival = 1;
      bool  booly = 2;
    }
  }
}
```

## 更新消息类型

在不破坏任何现有代码的情况下更新消息类型非常简单，记住下面的规则：

* 不要修改已存在的字段编号

* 如果添加了新字段，那么使用旧消息格式的代码序列化的消息仍然可以由新生成的代码解析。你应该记住这些元素的默认值，以便新代码能够正确地与旧代码生成的消息交互。类似地，新代码创建的消息可以由旧代码解析:旧二进制文件在解析时简单地忽略新字段。

* 可以删除字段，只要字段号不在后边更新消息类型时再次被使用。相反，也可以重命名字段，添加前缀“OBSOLETE_”或者保留字段号，以便将来的.proto用户不会意外地重用该编号。

* int32、uint32、int64、uint64和bool都是兼容的——这意味着可以将字段从其中一种类型更改为另一种类型，而不破坏向前或向后兼容性。如果解析出一个不适合对应类型的数字，您将得到与在c++中强制转换该数字的效果相同的结果(例如，如果一个64位的数字被读取为int32，它将被截断为32位)。

* sint32和sint64彼此兼容，但不兼容其它整数类型。

* string和bytes是兼容的，只要字节是有效的UTF-8。

* Embedded messages are compatible with bytes if the bytes contain an encoded version of the message.

* fixed32与sfixed32兼容，fixed64与sfixed64兼容。

* For string, bytes, and message fields, optional is compatible with repeated. Given serialized data of a repeated field as input, clients that expect this field to be optional will take the last input value if it's a primitive type field or merge all input elements if it's a message type field. Note that this is not generally safe for numeric types, including bools and enums. Repeated fields of numeric types can be serialized in the packed format, which will not be parsed correctly when an optional field is expected.

* enum is compatible with int32, uint32, int64, and uint64 in terms of wire format (note that values will be truncated if they don't fit). However be aware that client code may treat them differently when the message is deserialized: for example, unrecognized proto3 enum types will be preserved in the message, but how this is represented when the message is deserialized is language-dependent. Int fields always just preserve their value.

* Changing a single value into a member of a new oneof is safe and binary compatible. Moving multiple fields into a new oneof may be safe if you are sure that no code sets more than one at a time. Moving any fields into an existing oneof is not safe.

## 未知字段

未知字段是格式良好的pb序列化数据时，表示解析器无法识别的字段。例如，当旧二进制文件解析带有新字段的新二进制文件发送的数据时，这些新字段在旧二进制文件中成为未知字段。

最初，在解析过程中，proto3消息总是丢弃未知字段，但在3.5版本中，我们重新引入了保留未知字段的功能，以匹配proto2行为。在版本3.5及更高版本中，解析期间会保留未知字段，并包含在序列化输出中。


## Any

Any类型的消息不在.proto文件中定义，可以直接被嵌入使用。Any以字节的形式包含任意序列化的消息，以及充当该消息类型的全局惟一标识符并解析为该消息类型的URL。要使用Any类型，需要导入`import google/protobuf/any.proto`

```
import "google/protobuf/any.proto";

message ErrorStatus {
  string message = 1;
  repeated google.protobuf.Any details = 2;
}
```

给定消息类型的默认类型URL为:`type.googleapis.com/_packagename_._messagename_`

不同的语言实现将支持运行时库以类型安全的方式打包和解包Any消息


## Oneof

如果一个包含多个字段的消息，并且最多只会同时使用其中的一个字段，那么就可以通过使用OneOf特性来节省内存。

除了OneOf中的所有字段共享内存之外，OneOf中的字段类与常规字段一样。最多只能使用OneOf中的一个字段，设置OneOf中的任何一个的字段其它成员将自动被清除。根据选择的语言，可以使用特殊的case()或WhichOneof()方法检查oneof中的哪个字段被使用了(如果有的话)。

**使用OneOf**

```
message SampleMessage {
  oneof test_oneof {
    string name = 4;
    SubMessage sub_message = 9;
  }
}
```

可以添加任何类型的字段到OneOf定义中,除了map和repeated类型字段

在生成的代码中，其中OneOf中的字段具有与常规字段相同的getter和setter。您还可以得到一个特殊的方法来检查OneOf使用了哪个字段(如果有的话)

**Oneof的特点**

* 使用了其中的一个字段将会清除其它字段

* 如果解析器发现Oneof中的多个成员，只会保留最后一个

* A oneof cannot be repeated.

* Reflection APIs work for oneof fields

* ...


## Map

```
map<key_type, value_type> map_field = N;
```

key_type可以是任何正数或字符串类型(除了浮点类型和字节之外的任何标量类型)，不能是枚举类型；value_type可以是除其它map之外的任何类型。

* Map fields cannot be repeated

* map中元素的顺序是不固定的，不能依赖顺序

* When generating text format for a .proto, maps are sorted by key. Numeric keys are sorted numerically.

* 如果遇到重复的key则使用最新一个覆盖前一个

* 如果只有key没有值，根据不用的语言可能在序列化时会忽略这个key

**向后兼容性**

map语法等价于以下声明，所以pb实现不支持映射仍然可以处理你的数据:

```
message MapFieldEntry {
  key_type key = 1;
  value_type value = 2;
}

repeated MapFieldEntry map_field = N;
```

任何支持map的protocol buffers实现都必须生成和接受上述定义的格式。


## Packages

在.proto文件中添加一个可选的package说明符,可以防止协议消息类型之间的名称冲突。

```
package foo.bar;
message Open { ... }
```

使用对应的类型时需要指明package

```
message Foo {
  ...
  foo.bar.Open open = 1;
  ...
}
```

package说明符影响生成代码的方式取决于所选的语言:

* Java，package转为对应的java package,除非显示指定了option java_package

* Go，package转为Go的包名，除非显示指定了option go_package

* C#，package转为命名空间，除非显示指定了option csharp_namespace

* ...


## Defining Services

如果你想在RPC中使用定义的消息类型，可以在.proto文件中定义一个RPC service接口,pb编译器将根据选择的语言生成对应的定义接口和stub代码。

```
service SearchService {
  rpc Search(SearchRequest) returns (SearchResponse);
}
```

## JSON Mapping

Proto3支持JSON格式的规范编码，这使得在系统之间共享数据变得更加容易。编码在下表中按类型逐个描述。

如果json编码的数据中缺少一个值，或者它的值为null，那么在解析到pb时，它将被解释为适当的默认值。如果一个字段在pb中有默认值，那么默认情况下它将在json编码的数据中被省略以节省空间。具体的实现可以提供选项以在json编码的输出中包含具有默认值的字段。


| proto3                 | JSON          | JSON example                            | 备注 |
| ---------------------- | ------------- | --------------------------------------- | ---- |
| message                | object        | {"fooBar": v, "g": null, …}             |      |
| enum                   | string        | "FOO_BAR"                               |      |
| map\<K,V>               | object        | {"k": v, …}                             |      |
| repeated V             | array         | [v, …]                                  |      |
| bool                   | true, false   | true, false                             |      |
| string                 | string        | "Hello World!"                          |      |
| bytes                  | base64 string | "YWJjMTIzIT8kKiYoKSctPUB+"              |      |
| int32, fixed32, uint32 | number        | 1, -10, 0                               |      |
| int64, fixed64, uint64 | string        | `"1", "-10"                             |      |
| float, double          | number        | 1.1, -10.0, 0, "NaN", "Infinity"        |      |
| Any                    | object        | {"@type": "url", "f": v, … }            |      |
| Timestamp              | string        | "1972-01-01T10:00:20.021Z"              |      |
| Duration               | string        | "1.000340012s", "1s"                    |      |
| Struct                 | object        | { … }                                   |      |
| Wrapper types          | various types | 2, "2", "foo", true, "true", null, 0, … |      |
| FieldMask              | string        | "f.fooBar,h"                            |      |
| ListValue              | array         | [foo, bar, …]                           |      |
| Value                  | value         |                                         |      |
| NullValue              | null          |                                         |      |
| Empty                  | object        | {}                                      |      |




## Options

.proto文件中的单个声明可以用许多option进行注释,option不会改变声明的总体含义，但可能会影响在特定上下文中处理声明的方式。可用的option定义在`google/protobuf/descriptor.proto`

一些选项是file级选项，这意味着它们应该在顶级作用域中编写，而不是在任何message、enum或service定义中。有些选项是message级选项，这意味着它们应该写在message的定义中。有些选项是field级选项，这意味着它们应该写入字段定义中。选项也可以用于枚举类型、枚举值、oneof字段、service类型和service方法，但是，目前还没有任何有用的选项。

下面是一些最常用的选项:

* java_package(文件级别)：生成Java类要使用的包。如果.proto文件中没有给出显式的java_package选项，那么默认情况下将使用proto包(.proto文件中的“package”关键字指定)。然而，proto包通常不能成为好的Java包，因为proto包不一定是以反向域名开始。如果不生成Java代码，则此选项无效。

```
option java_package = "com.example.foo";
```

* java_multiple_files(文件级别)：将顶级的message、enums和service定义在不同的文件中，而不是定义在以.proto文件命名的外部类中。如果不生成Java代码，则此选项无效。

```
option java_multiple_files = true;
```

* java_outer_classname(文件级别)：生成的类的名字，如果没有`.proto`文件中没有明确指定java_outer_classname，那么将会在.proto文件的名字转为类的名字。如果不生成Java代码，则此选项无效。

```
option java_multiple_files = true;
```

* optimize_for(文件级别)：可以设置为SPEED、CODE_SIZE或LITE_RUNTIME。这会对c++和Java代码生成器(可能还有第三方生成器)产生以下影响:

    * ...

    * ...

```
option optimize_for = CODE_SIZE;
```   

* cc_enable_arenas(文件级别)：为C++启用arena allocation
 
* objc_class_prefix(文件级别)：为生成的Objective-C类设置前缀

* deprecated(字段级别)：如果设置为真，则指示该字段不赞成使用，并且不应由新代码使用。在大多数语言中，这没有实际效果。在Java中，这变成了@Deprecated注释。将来，其他特定于语言的代码生成器可能会在字段的访问器上生成deprecation注释，这将导致在编译试图使用该字段的代码时发出警告。如果没有人使用该字段，并且您希望阻止新用户使用该字段，可以考虑使用保留语句(reserved)替换该字段声明。

**自定义选项**

Protocol Buffers还允许定义和使用自己的选项。这是一个大多数人不需要的高级特性。如果你确实认为需要创建自己的选项，请参阅Proto2语言指南了解详细信息。注意创建自定义选项使用扩展，这只允许在proto3中自定义选项。

## 代码生成

要生成Java、Python、c++、Go、Ruby、Objective-C或c#代码，需要在.proto文件中定义的message类型,然后使用pb编译器.protoc文件。如果还没安装，需要先安装[pb编译器](https://developers.google.com/protocol-buffers/docs/downloads),然后按照README中的步骤进行操作。对于Go，还需要为编译器安装一个特殊的代码生成器[插件]。(https://github.com/golang/protobuf/)
  
Protocol编译器的调用如下:

`protoc --proto_path=_IMPORT_PATH_ --cpp_out=_DST_DIR_ --java_out=_DST_DIR_ --python_out=_DST_DIR_ --go_out=_DST_DIR_ --ruby_out=_DST_DIR_ --objc_out=_DST_DIR_ --csharp_out=_DST_DIR_ _path/to/file_.proto`

* IMPORT_PATH：指定解析`import`指令时查找.proto文件的目录，如果省略则使用当前目录。如果需要指定多个搜索目录，需要指定--proto_path选项多次。-I=_IMPORT_PATH_是--proto_path的简写。

* 可以指定多个输出指令

  * --java_out 在指定的DST_DIR目录中生成Java代码

  * --go_out  在指定的DST_DIR目录中生成Go代码

  * --csharp_out 在指定的DST_DIR目录中生成C#代码

* 必须提供一个或多个.proto文件作为输入,可以同时指定多个.proto文件。