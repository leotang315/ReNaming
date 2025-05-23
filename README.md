# 批量文件重命名工具

一个强大的命令行工具，用于批量重命名文件，支持多种占位符和自定义格式。

## 功能特性
- 多种占位符替换：
  - `{name}` 原文件名（不含扩展名）
  - `{lower}` 原文件名小写
  - `{upper}` 原文件名大写
  - `{ext}` 文件扩展名
  - `{date}` 当前日期
  - `{time}` 当前时间
  - `{datetime}` 日期+时间
  - `{index:<起始值>:<补零位数>}` 自动递增索引
  - `{regex:expr:group}` 正则表达式提取（group 0表示整个匹配字符串，1表示第一个匹配组）
  - `{split:sep:index}` 字符串分割（index从0开始）
  - `{slice:start:end}` 字符串切片 （从0字符开始，包含start，不包含end）
  - `{replace:old:new}` 字符串替换
- 双阶段安全重命名：
  ```bash
  # 第一阶段：生成重命名映射文件
  ReNaming -f .\photos -p "*.jpg" -n "vacation_{index}.jpg" -g plan.txt
  
  # 手动验证/修改映射文件后

  # 第二阶段：应用安全重命名
  ReNaming -a plan.txt

## 配置选项

| 参数        | 类型   | 必需 | 默认值       | 说明                                  | 示例                      |
|-------------|--------|------|--------------|---------------------------------------|---------------------------|
| `-i`        | 字符串 | 是*  | -            | 输入目录路径                          | `-i "D:\照片"`           |
| `-o`        | 字符串 | 是*  | -            | 新文件名模板                          | `-o "图片_{index}.{ext}"` |
| `-f`        | 字符串 | 否   | `*`          | 文件匹配模式                          | `-f "*.jpg"`             |
| `-df`       | 字符串 | 否   | YYYY-MM-DD   | 日期格式                              | `-df "YYYYMMDD"`         |
| `-tf`       | 字符串 | 否   | HH:MM:SS     | 时间格式                              | `-tf "HHmm"`             |
| `-g`        | 字符串 | 否   | -            | 生成映射文件                          | `-g plan.txt`            |
| `-a`        | 字符串 | 否   | -            | 应用映射文件                          | `-a plan.txt`             |

## 使用示例
### 基本使用
```bash
#输入文件：`IMG001.jpg` → 输出文件：`vacation_1_2023-10-01.jpg`
ReNaming -i ./photos -f "*.jpg" -o "vacation_{index}_{date}.{ext}"
```

### 映射文件使用
```bash
#生成映射文件（试运行）
ReNaming -i .\photos -f "*.jpg" -o "vacation_{index}.jpg" -g plan.csv  

#应用映射文件
ReNaming -a plan.csv 
```


### replace分隔符分割
```bash
#输入文件：`123_file.txt` → 输出文件：`abc_file.txt`
ReNaming -i ./files -o "{replace:123:abc}.{ext}"
```


### split分隔符分割
```bash
#输入文件：`123_file.txt` → 输出文件：`123.txt`
ReNaming -i ./files -o "{split:_:0}.{ext}"
```


### slice字符串切片
```bash
#输入文件：`123_file.txt` → 输出文件：`file.txt`
ReNaming -i ./files -o "{slice:4:7}.{ext}"
```


### regex正则表达式
```bash
#输入文件：`123_file.txt` → 输出文件：`123.txt`
ReNaming -i ./files -o "{regex:(\d+)_file:1}.{ext}"
```



## 使用注意事项
1. 处理顺序：文件按文件系统自然顺序处理（非确定性排序）
2. 特殊字符：模板中避免使用 `<>:"/\|?*` 等文件系统保留字符
3. 索引重置：每次程序运行后索引计数器会自动重置

## License
MIT