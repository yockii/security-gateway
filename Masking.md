# 脱敏

- 配置服务的人员信息获取接口，以便网关拦截并获取人员匹配数据
- 配置人员通用密级，以数字表示，数字越大，密级越高
- 配置人员在服务中的密级，以数字表示，数字越大，密级越高，若未配置，则将使用通用密级
- 配置通用字段的密级及脱敏规则，以数字表示，数字越大，密级越高，通用字段将尝试脱敏所有返回信息中的该字段
- 配置服务字段的密级及脱敏规则，以数字表示，数字越大，密级越高，服务字段将尝试脱敏所有该服务返回信息中的该字段
- 配置路由字段的密级及脱敏规则，以数字表示，数字越大，密级越高，路由字段将尝试脱敏所有该路由返回信息中的该字段

## 脱敏级别规则

脱敏级别从1-4共四个级别，对应的规则配置为：脱敏类型-[字符标记(可选)]*****  如下：

```
(all|each|start|middle|end)-[^](*)
```

- all: 全部替换为`-`后面配置的字符
- each: 每个字符脱敏为`-`后面配置的字符，如果配置了`^`则根据后面配置的字符数量来表示每N个字替换为一个脱敏字符
- start: 从头开始脱敏，脱敏字符数量根据后面配置的字符数量来表示
- middle: 从中间开始脱敏，脱敏字符数量根据后面配置的字符数量来表示
- end: 从尾部开始脱敏，脱敏字符数量根据后面配置的字符数量来表示
- `^`: 用于标记每N个字符替换为一个脱敏字符（取第一个脱敏字符），N为后面配置的脱敏字符数
- `*`: 用于标记脱敏字符，可以是任意字符，这里用*表示

## 脱敏规则示例

- all-*****: 全部替换为`*****`
- each-*****: 每个字符替换为`*****`，如`abc`替换为`***************`
- each-^***: 每3个字符替换为`*`，如`abcdef`替换为`**`
- start-***: 从头开始脱敏，脱敏3个字符，如`abcdef`替换为`***def`
- middle-***: 从中间开始脱敏，脱敏3个字符，如`abcdef`替换为`ab***f`
- end-***: 从尾部开始脱敏，脱敏3个字符，如`abcdef`替换为`abc***`

注：start、middle、end规则中，^不起作用，同时，如果字符数量不足，则全部脱敏

# 脱敏说明

所有请求默认密级1，会匹配密级1的脱敏规则，如果没有匹配到，则不脱敏

密级优先级：

1. 配置的用户-服务对应密级
2. 配置的用户密级
3. 密级1

