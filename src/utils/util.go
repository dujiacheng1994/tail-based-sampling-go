package utils

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
)

var DataSourcePort string

func MD5(key string) string{
	byteKey := []byte(key)
	md5Ctx := md5.New()
	md5Ctx.Write(byteKey)
	cipherStr := md5Ctx.Sum(nil)

	file, _ := os.OpenFile("md5.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666) //打开日志文件，不存在则创建
	defer file.Close()
	fmt.Fprintln(file,cipherStr)
	fmt.Fprintln(file,md5.Sum(byteKey))
	fmt.Fprintf(file,"%x\n", cipherStr)
	fmt.Fprintln(file,hex.EncodeToString(cipherStr))
	//fmt.Println(cipherStr)
	//fmt.Printf("%x\n", md5.Sum(byteKey))
	//fmt.Printf("%x\n", cipherStr)
	//fmt.Println(hex.EncodeToString(cipherStr))

	return hex.EncodeToString(cipherStr)
}

//     public static String MD5(String key) {
//         char hexDigits[] = {
//                 '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'A', 'B', 'C', 'D', 'E', 'F'
//         };
//         try {
//             byte[] btInput = key.getBytes();
//             // 获得MD5摘要算法的 MessageDigest 对象
//             MessageDigest mdInst = MessageDigest.getInstance("MD5");
//             // 使用指定的字节更新摘要
//             mdInst.update(btInput);
//             // 获得密文
//             byte[] md = mdInst.digest();
//             // 把密文转换成十六进制的字符串形式
//             int j = md.length;
//             char str[] = new char[j * 2];
//             int k = 0;
//             for (int i = 0; i < j; i++) {
//                 byte byte0 = md[i];
//                 str[k++] = hexDigits[byte0 >>> 4 & 0xf];
//                 str[k++] = hexDigits[byte0 & 0xf];
//             }
//             return new String(str);
//         } catch (Exception e) {
//             return null;
//         }
//     }