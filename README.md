# 批量文件重命名工具

一个强大的命令行工具，用于批量重命名文件，支持多种占位符和自定义格式。

## 功能特性

- 支持递归处理子目录
- 多种占位符替换：
  - `{name}` 原文件名（不含扩展名）
  - `{ext}` 文件扩展名
  - `{date}` 当前日期
  - `{time}` 当前时间
  - `{datetime}` 日期+时间
  - `{index}` 自动递增序号
  - `{regex:expr:group}` 正则表达式提取
  - `{split:sep:index}` 分隔符分割
  - `{slice:start:end}` 字符串切片
  - `{replace:old:new}` 字符串替换
  - 大小写转换：`{lower}`, `{upper}`, `{title}`
- 干运行模式（预览更改）
- 双阶段安全重命名：
  ```bash
  # 第一阶段：生成重命名映射文件
  ReNaming -i .\photos -p "*.jpg" -n "vacation_{index}.jpg" -g plan.txt
  
  # 手动验证/修改映射文件后

  # 第二阶段：应用安全重命名
  ReNaming -a plan.txt

## 配置选项

| 参数        | 类型   | 必需 | 默认值       | 说明                                  | 示例                      |
|-------------|--------|------|--------------|---------------------------------------|---------------------------|
| `-f`        | 字符串 | 是*  | -            | 输入目录路径                          | `-i "D:\照片"`           |
| `-n`        | 字符串 | 是*  | -            | 新文件名模板                          | `-n "图片_{index}.{ext}"` |
| `-p`        | 字符串 | 否   | `*`          | 文件匹配模式                          | `-p "*.jpg"`             |
| `-r`        | 布尔   | 否   | false        | 递归处理子目录                        | `-r`                     |
| `-df`       | 字符串 | 否   | YYYY-MM-DD   | 日期格式                              | `-df "YYYYMMDD"`         |
| `-tf`       | 字符串 | 否   | HH:MM:SS     | 时间格式                              | `-tf "HHmm"`             |
| `-d`        | 布尔   | 否   | false        | 干运行模式                            | `-d`                     |
| `-g`        | 字符串 | 否   | -            | 生成映射文件                          | `-g plan.txt`            |
| `-a`        | 字符串 | 否   | -            | 应用映射文件                          | `-a plan.txt`             |
## 使用示例

### 基本使用
```bash
ReNaming -i ./photos -p "*.jpg" -n "vacation_{index}_{date}.{ext}"
```
输入文件：`IMG001.jpg` → 输出文件：`vacation_1_2023-10-01.jpg`

### 递归处理
```bash
ReNaming -i ./documents -r -n "doc_{index}_{time}.{ext}"
```

### 自定义格式
```bash
ReNaming -i ./logs -df "20060102" -tf "150405" -n "log_{datetime}.txt"
```

### 正则表达式
```bash
ReNaming -i ./files -n "{regex:(\d+)_file:1}.{ext}"
```
输入文件：`123_file.txt` → 输出文件：`123.txt`

### 冲突处理
```bash
ReNaming -i ./data -c rename -n "dataset_{index}.csv"
```


## License
MIT