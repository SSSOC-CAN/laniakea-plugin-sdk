# Laniakea Plugin SDK
Repository for Laniakea Plugin SDKs in various coding languages

# Examples

Examples in Go and Python can be found in the examples repository

# Python

To create python plugins, install the following dependencies
```
$ pip install grpcio grpcio-tools grpcio-health-checking laniakea-plugin-sdk pyinstaller
```
Once you've created a plugin, you must create an executable in order for Laniakea to run the plugin
```
$ pyinstaller -F plugin.py
```