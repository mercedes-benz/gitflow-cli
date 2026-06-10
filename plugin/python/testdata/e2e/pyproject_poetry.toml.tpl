[tool.poetry]
name = "test-python-project"
version = "{{.Version}}"
description = "Test Python project"
authors = ["Test Author <test@example.com>"]

[tool.poetry.dependencies]
python = ">=3.8"

[build-system]
requires = ["poetry-core"]
build-backend = "poetry.core.masonry.api"
