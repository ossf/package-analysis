#!/bin/bash

cd dashboard
npm run build

gsutil -h "Cache-Control:no-cache,max-age=0" cp -r dist/* gs://ossf-malware-analysis/
