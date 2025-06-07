---
name: Bug report
about: Create a report to help us improve
title: '[BUG] '
labels: 'bug'
assignees: 'fboukezzoula'
---

**Describe the bug**
A clear and concise description of what the bug is.

**To Reproduce**
Steps to reproduce the behavior:
1. Run command '...'
2. With options '...'
3. See error

**Expected behavior**
A clear and concise description of what you expected to happen.

**Environment:**
- OS: [e.g. Linux, macOS, Windows]
- Architecture: [e.g. amd64, arm64]
- XKS Version: [e.g. v1.0.0]
- Azure CLI Version: [e.g. 2.50.0]
- Go Version (if building from source): [e.g. 1.21.0]

**Configuration:**
```bash
# Your .env configuration (remove sensitive data)
AZURE_TENANTID=xxx
AZURE_APPID=xxx
# ... (without secrets)
```

**Command executed:**
```bash
xks kubectl get pods -A
```

**Error output:**
```
Error: authentication failed
```

**Additional context**
Add any other context about the problem here.

**Logs**
If applicable, add logs with `--debug --verbose` flags:
```
xks --debug --verbose kubectl get pods
```