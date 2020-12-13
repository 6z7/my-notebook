# ASN.1

ASN.1（Abstract Syntax Notation One) 是一套标准，是描述数据的表示、编码、传输、解码的灵活的记法。一种标准的接口描述语言(IDL)，用于定义可以跨平台序列化和反序列化的数据结构

ASN.1本身只定义了表示信息的抽象句法，但是没有限定其编码的方法。各种ASN.1编码规则提供了由ASN.1描述其抽象句法的数据的值的传送语法（具体表达）。标准的ASN.1编码规则有基本编码规则（BER，Basic Encoding Rules）、规范编码规则（CER，Canonical Encoding Rules）、唯一编码规则（DER，Distinguished Encoding Rules）、压缩编码规则（PER，Packed Encoding Rules）和XML编码规则（XER，XML Encoding Rules）。为了使ASN.1能够描述一些原先没有使用ASN.1定义，因此不适用上述任一编码规则的数据传输和表示的应用和协议，另外制订了ECN来扩展ASN.1的编码形式。ECN可以提供非常灵活的表明方法，但还没有得到普遍应用。