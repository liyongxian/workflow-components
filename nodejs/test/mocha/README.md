## 组件名称：Nodejs Mocha Test

### mocha：

使用 mocha 进行NodeJS的单元测试

### 组件参数
#### 入参
- `GIT_CLONE_URL` 必填，源代码地址，如为私有仓库需要授权; 如需使用系统关联的git仓库, 可以从系统提供的全局环境变量中获取: `${_WORKFLOW_GIT_CLONE_URL}`
- `GIT_REF` 非必填，源代码git目标引用，可以是一个git branch, git tag 或者git commit ID, 默认值master
- `TEST_PATH` 必填，目标文件路径
- `TEST_PARAMS` 非必填，如 `--timeout 3000`

#### 出参
- 无

### Tag列表及其Dockerfile链接

* 5.2, latest: [Dockerfile](https://github.com/tencentyun/workflow-components/blob/c587a3a7ba3632ab7422d2a08efd8bc60c93f5d2/nodejs/test/mocha/Dockerfile)

### 源码地址

Nodejs Mocha Test: <https://github.com/tencentyun/workflow-components/tree/master/nodejs/test/mocha>

### 构建:

`docker build -t hub.tencentyun.com/tencenthub/nodejs_test_mocha .`
