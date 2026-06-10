from setuptools import setup

setup(
    name="test-python-project",
    version="{{ .Version }}",
    description="Test Python project",
    author="Test Author",
    author_email="test@example.com",
    py_modules=["mymodule"],
    python_requires=">=3.8",
)