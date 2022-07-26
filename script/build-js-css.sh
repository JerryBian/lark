SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

mkdir -p $SCRIPT_DIR/../static/style/fonts
cp $SCRIPT_DIR/../node_modules/bootstrap-icons/font/fonts/* $SCRIPT_DIR/../static/style/fonts
echo "copy bootstrap icons fonts completed."

npx sass $SCRIPT_DIR/../static/style/style.scss $SCRIPT_DIR/../static/style/style.css
echo "sass completed."

npx uglifycss --ugly-comments --output $SCRIPT_DIR/../static/style/style.min.css $SCRIPT_DIR/../static/style/style.css
echo "css completed."

npx uglifyjs --compress -o $SCRIPT_DIR/../static/script/script.min.js \
    $SCRIPT_DIR/../node_modules/bootstrap/dist/js/bootstrap.js \
    $SCRIPT_DIR/../static/script/script.js
echo "js completed."