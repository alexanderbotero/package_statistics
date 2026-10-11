# Time

8 hours

# How I tackled this problem

The first thing that I did after reading the requirement, was to open one of the links and check the files, here I spent around and hour, what I was trying to find any gotcha, I was scanning the files, and try to see if I missed something, the only thing that I found was this text "Clients should correctly handle file names containing white space characters (possibly taking advantage of the fact that package names cannot include white space characters)." that make me think, space should not be the way to define each column, or at least the first one, after thinking I thouht to myself, do it backwards, and problem resolved, I ran this command to test my theory
```bash
rev Contents-udeb-s390x  | cut -d' ' -f1 | rev
``` 
I did that with a couple of those files, and when I noticed all was good, I decided that was the way to go.


after that I coded the solution, I downloaded the file, I was doing some research how to do it in the most efficient way, I found this one, https://www.luisllamas.es/lectura-escritura-ficheros-go-os-io/#lectura-incremental-con-bufio (Sorry, it is Spanish), with that I tested downloading the file and showing the info, it took me as well around an hour.

The next step was to create a first version of the cli, with a hardcoded arch, I got the result in the screen, and I tested again, I used this command to check some of the results.

```bash
grep -o "debian-installer/console-keymaps-usb" Contents-udeb-all | wc -l
```

At that moment I had a basic and funtional version of the assesment, I started to refectored it, and unit tested, fist I make the part of the parsing as separated funtion and tested it, secondly, the sorting, the third one was the dowloading, when I got this point, the code was most reable, this took me around 1 hour in a half.

I reread the code again, and I noticed a problem with the architecture param, I was not validating the input, I went again to the debian pages and analized the architectures, and I decided to use and regex to validate the format, I was fixing two problems here, first to have a valid architecture, and secondly a security one, someone could do something innapropiated with that parameter, I added the validation and the test. that took me 45 mins

finally, I was reading the assestment again, I saw the part of the mirrors, initially, the code was using only the first one that I found in the email, I decided to add the three of them and also adding some randomez, that way the code will not depend on one of those, I did not add testing for that part, due the nature of the randomnez, I mean, you can test something that is deterministc, that part is not, maybe I could isolate that part, but at that moment, for me it was overengeenering, finally I make the part of the top not fixed to 10, but I created a const, and make the adjustment, in the code and the tests. here I spent around 1 and a half hours more.


I took a break, trying to refresh my mind, and I came back 45 mins later, I commented the code, reviewed it again, I reread the assesment to see if I missed something, I spend 30 mins, and I staring writting down the readme.