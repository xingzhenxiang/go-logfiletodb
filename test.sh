for ((i=1;i<5000000;i++))
do 
echo $i ` date +%F_%T` ` head  /dev/urandom  |  tr -dc 0-9  | head -c 20 |md5sum`
sleep 1
done
