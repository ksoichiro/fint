#!/bin/sh

REPO_DIR=_master
REPORT_DIR=report
REPORT_DEFAULT_DIR=${REPORT_DIR}/default
REPORT_DARK_DIR=${REPORT_DIR}/dark
REPORT_AUTOFIX_DIR=${REPORT_DIR}/auto-fix

# Clone and create index.md
if [ ! -d ${REPO_DIR} ]; then
  git clone git@github.com:ksoichiro/fint.git ${REPO_DIR}
fi
pushd ${REPO_DIR} > /dev/null
git reset --hard HEAD
git checkout master
git fetch
git pull origin master
echo "---" > ../index.md
echo "layout: default" >> ../index.md
echo "---" >> ../index.md
cat README.md >> ../index.md
popd > /dev/null

# Create sample report
pushd ${REPO_DIR} > /dev/null
if [ -d ../${REPORT_DIR} ]; then
  rm -rf ../${REPORT_DIR}
fi

mkdir -p ../${REPORT_DEFAULT_DIR}
fint run -s testdata/objc/FintExample -i objc -h ../${REPORT_DEFAULT_DIR} -f -q

mkdir -p ../${REPORT_DARK_DIR}
fint run -s testdata/objc/FintExample -i objc -h ../${REPORT_DARK_DIR} -template dark -f -q

mkdir -p ../${REPORT_DEFAULT_DIR}
fint run -s testdata/objc/FintExample -i objc -h ../${REPORT_AUTOFIX_DIR} -f -fix -q

popd > /dev/null
