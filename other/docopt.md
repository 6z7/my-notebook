[docopt](http://docopt.org/)
---
命令行接口描述语言

docopt是基于几十年来的帮助信息和程序手册中使用的预定。docopt中的接口描述就是这样一个帮助消息，但是进行了格式化。如下:

        Naval Fate.

        Usage:
        naval_fate ship new <name>...
        naval_fate ship <name> move <x> <y> [--speed=<kn>]
        naval_fate ship shoot <x> <y>
        naval_fate mine (set|remove) <x> <y> [--moored|--drifting]
        naval_fate -h | --help
        naval_fate --version

        Options:
        -h --help     Show this screen.
        --version     Show version.
        --speed=<kn>  Speed in knots [default: 10].
        --moored      Moored (anchored) mine.
        --drifting    Drifting mine.


这个例子定义了一个可执行命令naval_fate，应用的名称叫做Naval Fate，包含不同的命令(ship, new, move等等),可选参数(-h, --help, --speed=\<kn>等等)，位置参数(\<name>, \<x>, \<y>)。

上边的例子中使用"[]"、"()"、"|"、"..."分别表示可选的、必选、互斥、重复元素的意思。

在Usage下边有一个可选项列表描述。它们秒数了是否选项具有长短格式(-h,--help)，是否选项有参数(--sped=\<kn>)，是否参数有默认值([default: 10])。

docopt实现将提取所有这些信息并生成一个命令行参数解析器，当使用-h或--help选项调用程序时，将显示为帮助消息。

## <center> Usage patterns

在关键字usage（不区分大小写）和可见空行之间出现的文本被解释为用法模式列表。usage后的第一个单词解析为程序的名称。下面是一个不需要命令行参数的程序的最小示例:

    Usage: my_program

程序可以有多个模式，其中列出了用于描述该模式的各种元素：

    Usage:
    my_program command --option <argument>
    my_program [<optional-argument>]
    my_program --another-option=<with-argument>
    my_program (--either-that-option | <or-this-argument>)
    my_program <repeating-argument> <repeating-argument>...

下面将描述每个元素和结构。我们将使用单词“word”来描述由空白、一个“[]（|）”字符或“…”字符分隔的字符序列。

### <center> \<argument> ARGUMENT </center>

以“<”开头、以“>”结尾的单词和大写单词被解释为位置参数。

    Usage: my_program <host> <port>

### <center> -o --option    </center>

以一个或两个破折号开头的单词（除了“-”、“--”本身）分别解释为短（一个字母）或长选项。

* 短选项可以堆叠，即使-abc与-a -b -c等价
* 长选项可以有参数通过"="和空格分割 --inout=ARG和--input ARG等价
* 短选项可以参数通过可选空格分割 -f FILE和-fFile等价

注意，--input ARG（与--input=ARG相反）是不明确的，这意味着无法区分ARG是选项的参数还是位置参数。在使用模式时，只有在提供了该选项的描述时，才会将其解释为带参数的选项。否则，它将被解释为一个选项和单独的位置参数。

对于-f文件和-fFILE也有同样的歧义，无法判断是多个堆叠的短选项还是带有参数的选项。只有在提供选项的说明时，这些符号才会解释为带参数的选项。

###  <center> command

所有其它不遵循上述--options或\<arguments>约定的单词都被解释为（子）命令。

###  <center> [optional elements]

被"[]"括起来的元素(options, arguments, commands)是可选的，元素是否包含在同一对或不同的括号中并不重要。如:

    Usage: my_program [command --option <argument>]
    等价
    Usage: my_program [command --option <argument>]


###  <center> (required elements)

默认情况下,未被"[]"括起来的元素都是必选的。然而,需要使用"()"将元素显示标记为必选。

例如，当需要对互斥元素进行分组时（请参见下一节）

    Usage: my_program (--either-this <and-that> | <or-this>)

另一种情况是当需要指定如果一个元素存在，则需要另一个元素时:

    Usage: my_program [(<one-argument> <another-argument>)]

在这种情况下，有效的程序调用可以没有参数，也可以有两个参数。

###  <center>  element|another

互斥元素可以用管道“|”分隔，如下所示：

    Usage: my_program go (--up | --down | --left | --right)

当需要相互排斥的情况之一时，使用"()"对元素进行分组。当可以不需要相互排斥的情况时，使用"[]"将元素分组：

    Usage: my_program go [--up | --down | --left | --right]

注意，多个模式的工作方式与管道“|”方式完全相同，即：

    Usage: my_program run [--fast]
        my_program jump [--high]
    等价    
    Usage: my_program (run [--fast] | jump [--high])


###  <center>  element...

使用省略号“…”指定左侧的参数（或参数组）可以重复一次或多次：

    Usage: my_program open <file>...
        my_program move (<from> <to>)...

可以灵活地指定所需参数的数目。以下是要求零个或多个参数的3种（冗余）方法：

    Usage: my_program [<file>...]
        my_program [<file>]...
        my_program [<file> [<file> ...]]

一个或多个参数：

    Usage: my_program <file>...

两个或多个参数（等等）：

    Usage: my_program <file> <file>...

### <center>  [options]

"[options]"是一种快捷方式，避免列出模式中的所有选项（从带有说明的选项列表中）。如:

    Usage: my_program [options] <path>

    --all             List everything.
    --long            Long output.
    --human-readable  Display in human-readable format.

等价

    Usage: my_program [--all --long --human-readable] <path>

    --all             List everything.
    --long            Long output.
    --human-readable  Display in human-readable format.

如果有许多选项，并且所有选项都适用于其中一个模式，那么这将非常有用。或者，如果有选项的短版本和长版本（在选项描述部分中指定），则可以在模式中列出它们中的任何一个：

    Usage: my_program [-alh] <path>

    -a, --all             List everything.
    -l, --long            Long output.
    -h, --human-readable  Display in human-readable format.

### <center> [--]

当“--”不是选项的一部分时，通常用作选项和位置参数的分隔符。以处理文件名为例可能被误认为选项的情况。为了支持这种约定，在位置参数之前在模式中添加“[--]”。

    Usage: my_program [options] [--] <file>...

除此特殊含义外，“--”只是一个普通命令，因此可以应用任何先前描述的操作，例如，使其成为必需的（通过删除"[]"）

### <center> [-]

如果不是选项的一部分，通常使用一个短划线“-”来表示程序应该处理stdin，而不是文件。如果要遵循此约定，请在模式中添加“[-]”。-“本身只是一个普通的命令，你可以用它来表达任何意义。


## <center> Option descriptions

选项描述由一个选项列表组成，这些选项放在usage下面。如果选项在usage中没有歧义，那么选项描述是可选的。

选项描述允许指定:

* 短选项和长选项是同义的
* 选项有参数
* 选项参数有默认值

规则如下:

以“-”或“--”（不包括空格）开头的每一行都被视为选项描述,如:

    Options:
    --verbose   # GOOD
    -o FILE     # GOOD
    Other: --bad  # BAD, line does not start with dash "-"

若要指定某个选项具有参数，请在空格（或等于“=”符号）后放置一个描述该参数的单词，如下所示,对于选项的参数，请遵循<尖括号>或大写约定。如果要分隔选项，可以使用逗号“,”。在下面的示例中，这两行都是有效的，但是建议使用单一风格。

    -o FILE --output=FILE       # without comma, with "=" sign
    -i <file>, --input <file>   # with comma, without "=" sign

使用两个空格分隔带有说明信息的选项。

    --verbose MORE text.    # BAD, will be treated as if verbose
                            # option had an argument MORE, so use
                            # 2 spaces instead
    -q        Quit.         # GOOD
    -o FILE   Output file.  # GOOD
    --stdout  Use stdout.   # GOOD, 2 spaces

如果要为带有参数的选项设置默认值，请将其以[default: \<the-default-value>]的形式放入选项的说明中。

    --coefficient=K  The K coefficient [default: 2.95]
    --output=FILE    Output file [default: test.txt]
    --directory=DIR  Some directory [default: ./]


## <center> [浏览器中尝试docopt](http://try.docopt.org/)

## <center> 实现

docopt在多种编程语言中都可以使用。官方实现见[docopt organization on GitHub.](https://github.com/docopt)


