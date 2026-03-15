#!/usr/bin/env python3
"""
Jinja2 渲染 Worker - 支持 JSON-RPC 通信

通过 stdin/stdout 与 Go 主程序通信，实现长驻进程模式
避免频繁启动 Python 进程的开销
"""

import sys
import json
import traceback
import logging
from jinja2 import Environment, FileSystemLoader, TemplateNotFound

# 导入自定义 filters
from filters import load_filters

# 导入 os 用于获取环境变量
import os


# 配置日志系统
def setup_logging():
    """配置日志记录器"""
    log_level = os.getenv("LOG_LEVEL", "INFO").upper()

    # 创建 logger
    logger = logging.getLogger("render_worker")
    logger.setLevel(getattr(logging, log_level, logging.INFO))

    # 创建 handler（输出到 stderr）
    handler = logging.StreamHandler(sys.stderr)
    handler.setLevel(getattr(logging, log_level, logging.INFO))

    # 创建 formatter
    formatter = logging.Formatter(
        "%(asctime)s [%(levelname)s] [%(name)s] %(message)s",
        datefmt="%Y-%m-%d %H:%M:%S",
    )
    handler.setFormatter(formatter)

    # 添加 handler 到 logger
    logger.addHandler(handler)

    return logger


class RenderWorker:
    """Jinja2 渲染 Worker"""

    def __init__(self, logger=None):
        """初始化 Jinja2 环境和 filters"""
        self.logger = logger or logging.getLogger("render_worker.worker")

        # 创建 Ansible 兼容的 Undefined 类
        # Ansible 在变量未定义时输出类信息：<class 'jinja2.utils.Namespace'>
        from jinja2 import Undefined

        class AnsibleUndefined(Undefined):
            """Ansible 兼容的 Undefined 类"""

            def __str__(self):
                return self._undefined_name

            def __repr__(self):
                return "jinja2.utils.Namespace"

        self.logger.info("初始化 Jinja2 环境")

        # 初始化 Jinja2 环境
        # 使用根目录加载器，支持绝对路径模板
        self.env = Environment(loader=FileSystemLoader("/"), undefined=AnsibleUndefined)

        # 加载自定义 filters
        filters = load_filters()
        self.env.filters.update(filters)
        self.logger.debug(f"已加载 {len(filters)} 个 filters")

        # 添加全局函数
        self.env.globals["range"] = range

        self.logger.info("Jinja2 环境初始化完成")

    def render(self, template_path: str, context: dict) -> str:
        """
        渲染模板

        Args:
            template_path: 模板文件绝对路径
            context: 渲染上下文

        Returns:
            渲染后的字符串

        Raises:
            Exception: 渲染失败时抛出异常
        """
        self.logger.debug(f"开始渲染模板：{template_path}")

        try:
            # 加载模板
            self.logger.debug(f"加载模板文件：{template_path}")
            template = self.env.get_template(template_path)

            # 渲染模板
            self.logger.debug(f"渲染模板，上下文变量数：{len(context)}")
            result = template.render(**context)

            self.logger.info(f"模板渲染成功：{template_path}")
            return result

        except TemplateNotFound:
            self.logger.error(f"模板文件不存在：{template_path}")
            raise Exception(f"模板文件不存在：{template_path}")
        except Exception as e:
            self.logger.error(
                f"渲染模板失败 [{template_path}]: {str(e)}", exc_info=True
            )
            raise Exception(f"渲染模板失败 [{template_path}]: {str(e)}")


def main():
    """主函数 - Worker 模式"""
    # 设置日志
    logger = setup_logging()
    logger.info("=" * 60)
    logger.info("Render Worker 启动")
    logger.info(f"Python 版本：{sys.version}")
    logger.info(f"工作目录：{sys.path}")
    logger.info("=" * 60)

    worker = RenderWorker(logger)

    # 检查是否以 worker 模式运行
    if len(sys.argv) > 1 and sys.argv[1] == "--worker-mode":
        logger.info("进入 Worker 模式（JSON-RPC over stdin/stdout）")

        # Worker 模式：持续从 stdin 读取请求
        request_count = 0
        error_count = 0

        while True:
            try:
                # 读取一行输入（JSON-RPC 请求）
                line = sys.stdin.readline()
                if not line:
                    logger.info("stdin 关闭，正常退出")
                    break

                request_count += 1
                logger.debug(f"收到请求 #{request_count}")

                # 解析请求
                try:
                    request = json.loads(line.strip())
                except json.JSONDecodeError as e:
                    logger.error(f"JSON 解析失败：{str(e)}")
                    response = {"success": False, "error": f"Invalid JSON: {str(e)}"}
                    print(json.dumps(response), flush=True)
                    error_count += 1
                    continue

                template_path = request.get("template_path")
                context = request.get("context", {})

                if not template_path:
                    logger.warning("请求缺少 template_path 参数")
                    response = {"success": False, "error": "template_path is required"}
                else:
                    logger.debug(f"渲染模板：{template_path}")
                    # 渲染模板
                    content = worker.render(template_path, context)
                    response = {"success": True, "content": content}
                    logger.info(f"请求 #{request_count} 处理成功")

                # 返回响应（JSON 格式）
                print(json.dumps(response), flush=True)

            except json.JSONDecodeError as e:
                # JSON 解析错误
                response = {"success": False, "error": f"Invalid JSON: {str(e)}"}
                print(json.dumps(response), flush=True)

            except Exception as e:
                # 其他异常
                error_count += 1
                logger.error(f"处理请求失败 #{request_count}: {str(e)}", exc_info=True)
                response = {
                    "success": False,
                    "error": str(e),
                    "traceback": traceback.format_exc(),
                }
                print(json.dumps(response), flush=True)

        # Worker 循环结束统计
        logger.info("=" * 60)
        logger.info("Worker 停止运行")
        logger.info(f"总处理请求数：{request_count}")
        logger.info(f"错误请求数：{error_count}")
        logger.info(
            f"成功率：{(request_count - error_count) / request_count * 100:.2f}%"
            if request_count > 0
            else "无请求"
        )
        logger.info("=" * 60)

    else:
        # 命令行测试模式
        logger.info("进入测试模式（命令行单次渲染）")

        if len(sys.argv) < 3:
            print("Usage:")
            print(
                "  Test mode: python3 render_worker.py <template_path> <context_json>"
            )
            print("  Worker mode: python3 render_worker.py --worker-mode")
            sys.exit(1)

        template_path = sys.argv[1]
        context_json = sys.argv[2]

        try:
            logger.info(f"测试模式 - 渲染模板：{template_path}")
            context = json.loads(context_json)
            content = worker.render(template_path, context)
            print(content)
            logger.info("测试模式渲染成功")
        except Exception as e:
            logger.error(f"测试模式渲染失败：{str(e)}", exc_info=True)
            print(f"Error: {e}", file=sys.stderr)
            sys.exit(1)


if __name__ == "__main__":
    main()
