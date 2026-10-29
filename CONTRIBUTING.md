# Contributing

感谢关注 **bay**。欢迎 Issue 与 Pull Request。

## 开发环境

- Go 1.22+（仓库按本机 toolchain 锁定；CI 使用稳定版本）
- 克隆后：

```bash
git clone git@github.com:CassianLi/bay.git
cd bay
make local
go test ./...
```

## 分支与 PR

1. 从 `main` 开分支（如 `feat/...`、`fix/...`）
2. 保持改动聚焦；附带或更新相关测试
3. PR 说明：动机、行为变化、如何验证
4. 确保 `go test ./...` 通过

## 代码约定

- CLI 子命令放在 `cmd/bay/`
- 领域逻辑放在 `internal/`
- 用户可见输出注意终端对齐（含 CJK 显示宽度）
- 不要提交本地登记库、密钥、真实项目路径样例中的隐私数据

## 文档

- 用户文档：根目录 `README.md`
- 设计与实现计划：`docs/`（见 `docs/README.md`）
- 行为变更时请同步更新 README 命令表

## License

贡献代码即表示同意以仓库根目录 [LICENSE](LICENSE)（MIT）授权。
