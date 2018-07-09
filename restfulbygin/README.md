# restfulbygin使用说明

---

[toc]

---

## 程序简介

本程序使用[gin](github.com/gin-gonic/gin)框架实现简单的Restful api服务，主要实现以下功能：

1. 用户鉴权，使用JWT实现用户鉴权
2. 用户分权，区分普通用户和管理员用户
3. 用户限速，针对不同的用户实行不同的限速策略
4. 用户增删改查

## 下载安装

    go get -u github.com/fourth04/restfulbygin

用户鉴权数据存储于数据库，所以需要安装数据库，本程序使用gorm作为orm库，支持SQLit、MySQL、postgres等

## 编译

进入`restfulbygin/server`目录，运行：

    go build

即可

## 快速使用

### 修改配置文件

配置文件模板存放于：`restfulbygin/docs/config.json`文件，修改相关项即可

### 使用命令运行服务

    .\server -c config.json

`注意`：初始运行程序时会自动新建users表，但是由于api做了鉴权，只有管理员才有权限增删改查用户，而此时数据库中并没有数据，所以直接使用sql，新建一个使用`admin/admin`作为账号密码的用户：

    INSERT INTO users("username", "password", "salt", "role_name", "rate_formatted") VALUES ('admin', 'ba8ad26b5fa6e20a651a5829d336c37c', 'e16PLcdEan', 'admin', '10000-M');

### api使用说明

以下演示使用[httpie](https://github.com/jkbrzt/httpie)实现，可自行使用curl，或者带界面的postman

#### 获取JWT

首先设置以下环境变量：

```bash
# Linux
export BASE_URL="http://localhost:8080"

# Windows
set BASE_URL="http://localhost:8080"
```

修改 username/password 以便获取JWT:

```bash
# Linux
http POST $BASE_URL/login grant_type=password username=admin password=123456

# Windows
http POST %BASE_URL%/login grant_type=password username=admin password=123456
```

正确响应样例：

```json
{
    "code": 200,
    "expire": "2018-07-08T13:01:55+08:00",
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJDcmVhdGVkQXQiOiIyMDE4LTA3LTA3VDEwOjQ4OjE3LjM1MTUwOCswODowMCIsIklEIjoxLCJSYXRlRm9ybWF0dGVkIjoiMTAwMDAtTSIsIlJvbGVOYW1lIjoiYWRtaW4iLCJVcGRhdGVkQXQiOiIyMDE4LTA3LTA3VDEwOjQ4OjE3LjM1MTUwOCswODowMCIsIlVzZXJuYW1lIjoiYWRtaW4iLCJleHAiOjE1MzEwMjYxMTUsImlkIjoiYWRtaW4iLCJvcmlnX2lhdCI6MTUzMTAxODkxNX0.Lt1ouXhPL3-IXCrhTzfuir-7fx0bkuRqQ8els0VAOnw"
}
```

#### 刷新JWT

首先设置以下环境变量：

```bash
# Linux
export JWT_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJDcmVhdGVkQXQiOiIyMDE4LTA3LTA3VDEwOjQ4OjE3LjM1MTUwOCswODowMCIsIklEIjoxLCJSYXRlRm9ybWF0dGVkIjoiMTAwMDAtTSIsIlJvbGVOYW1lIjoiYWRtaW4iLCJVcGRhdGVkQXQiOiIyMDE4LTA3LTA3VDEwOjQ4OjE3LjM1MTUwOCswODowMCIsIlVzZXJuYW1lIjoiYWRtaW4iLCJleHAiOjE1MzEwMjYxMTUsImlkIjoiYWRtaW4iLCJvcmlnX2lhdCI6MTUzMTAxODkxNX0.Lt1ouXhPL3-IXCrhTzfuir-7fx0bkuRqQ8els0VAOnw"

# Windows
set JWT_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJDcmVhdGVkQXQiOiIyMDE4LTA3LTA3VDEwOjQ4OjE3LjM1MTUwOCswODowMCIsIklEIjoxLCJSYXRlRm9ybWF0dGVkIjoiMTAwMDAtTSIsIlJvbGVOYW1lIjoiYWRtaW4iLCJVcGRhdGVkQXQiOiIyMDE4LTA3LTA3VDEwOjQ4OjE3LjM1MTUwOCswODowMCIsIlVzZXJuYW1lIjoiYWRtaW4iLCJleHAiOjE1MzEwMjYxMTUsImlkIjoiYWRtaW4iLCJvcmlnX2lhdCI6MTUzMTAxODkxNX0.Lt1ouXhPL3-IXCrhTzfuir-7fx0bkuRqQ8els0VAOnw"
```

```bash
# Linux
http POST $BASE_URL/auth/refresh_token Authorization:"Bearer $JWT_TOKEN"

# Windows
http GET %BASE_URL%/auth/refresh_token Authorization:"Bearer %JWT_TOKEN%"
```
#### 用户管理

##### 获取所有用户

```bash
# Linux
http $BASE_URL/api/users Authorization:"Bearer $JWT_TOKEN"

# Windows
http %BASE_URL%/api/users Authorization:"Bearer %JWT_TOKEN%"
```

##### 通过某个id获取用户

```bash
# Linux
http $BASE_URL/api/users/1 Authorization:"Bearer $JWT_TOKEN"

# Windows
http %BASE_URL%/api/users/1 Authorization:"Bearer %JWT_TOKEN%"
```

##### 新增用户

```bash
# Linux
http POST $BASE_URL/api/users Authorization:"Bearer $JWT_TOKEN" username=test password=123456 role_name=user rate_formatted=100-S

# Windows
http POST %BASE_URL%/api/users Authorization:"Bearer %JWT_TOKEN%" username=test password=123456 role_name=user rate_formatted=100-S
```
##### 更新用户

```bash
# Linux
http PUT $BASE_URL/api/users/2 Authorization:"Bearer $JWT_TOKEN" username=test_update password=654321 role_name=admin rate_formatted=1000-H

# Windows
http PUT %BASE_URL%/api/users/2 Authorization:"Bearer %JWT_TOKEN%" username=test_update password=654321 role_name=admin rate_formatted=1000-H
```

##### 删除用户

```bash
# Linux
http DELETE $BASE_URL/api/users/4 Authorization:"Bearer $JWT_TOKEN"

# Windows
http DELETE %BASE_URL%/api/users/4 Authorization:"Bearer %JWT_TOKEN%"
```
