# Python tus clients

There are two tus Python client libraries that were tested, ```tuspy``` and ```tus.py```.

## tuspy

Documentation: [tuspy](https://tus-py-client.readthedocs.io/en/latest/tusclient.html#module-tusclient.client)

### Test Steps
Install the tuspy Python module:
```bash
pip install tuspy
```

Run the tusclient uploader test:
```bash
pip install tuspy
python tuspy-uploader-test.py
```

## tus.py

Documentation: [tus.py](https://github.com/cenkalti/tus.py)

### Test Steps
Install the tuspy Python module:
```bash
pip install -U tus.py
```

Run the tus.pyt uploader test:
```bash
pip install tus
python tus-uploader-test.py
```

The ```tus.py``` upload step can also be run directly from the command line to monitor progress written to stdout of the terminal as it runs.
```bash
tus-upload 10MB-test-file https://as-bulk-upload-tusd.azurewebsites.net/files/
```