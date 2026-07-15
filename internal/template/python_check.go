package template

import (
	"fmt"
	"os/exec"
	"strings"
)

// PythonDepsCheck 检查 Python 依赖是否满足要求
// 返回友好报错，指导用户安装缺失的依赖
func PythonDepsCheck() error {
	// 1. 检查 python3 是否存在
	if err := checkPython3(); err != nil {
		return err
	}

	// 2. 检查必需的 Python 包
	if err := checkRequiredPackages(); err != nil {
		return err
	}

	return nil
}

// checkPython3 检查 python3 是否可用
func checkPython3() error {
	cmd := exec.Command("python3", "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf(`未找到 python3，请先安装 Python 3：

  macOS:
    brew install python3

  Ubuntu/Debian:
    sudo apt-get install python3 python3-pip

  Windows:
    从 https://www.python.org/downloads/ 下载安装

安装后确保 python3 在 PATH 中，然后重新运行 k8s-gen`)
	}

	// 提取版本号
	version := strings.TrimSpace(string(output))
	if !strings.HasPrefix(version, "Python 3") {
		return fmt.Errorf(`需要 Python 3.x，当前版本：%s

请安装 Python 3.8 或更高版本：https://www.python.org/downloads/`, version)
	}

	return nil
}

// checkRequiredPackages 检查必需的 Python 包
func checkRequiredPackages() error {
	// 需要检查的包及对应的导入名
	requiredPackages := []struct {
		packageName string // pip 包名
		importName  string // import 导入名
		description string // 用途说明
	}{
		{"Jinja2", "jinja2", "模板引擎"},
		{"PyYAML", "yaml", "YAML 处理"},
	}

	missing := []string{}

	for _, pkg := range requiredPackages {
		if !checkImport(pkg.importName) {
			missing = append(missing, pkg.packageName)
		}
	}

	if len(missing) > 0 {
		installCmd := "pip3 install " + strings.Join(missing, " ")
		return fmt.Errorf(`缺少 Python 依赖包：%s

请运行以下命令安装：

  %s

或者使用项目自带的 requirements.txt：

  pip3 install -r scripts/requirements.txt

安装完成后重新运行 k8s-gen`, strings.Join(missing, ", "), installCmd)
	}

	return nil
}

// checkImport 检查某个包是否可导入
func checkImport(importName string) bool {
	// 使用 python3 -c "import xxx" 检查包是否可导入
	cmd := exec.Command("python3", "-c", fmt.Sprintf("import %s; print('ok')", importName))
	err := cmd.Run()
	return err == nil
}
