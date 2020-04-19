tee xx.txt 重定向写入xx.txt文件中的同时也在终端显示

${BASH_SOURCE-$0}  获取当前执行的脚本文件的全路径


netstat 

stat

搜索结果中加上ps aux|head -n 1;ps -aux|grep senki


孤儿进程：一个父进程退出，而它的一个或多个子进程还在运行，那么那些子进程将成为孤儿进程。孤儿进程将被init进程(进程号为1)所收养，并由init进程对它们完成状态收集工作。　　

僵尸进程：一个进程使用fork创建子进程，如果子进程退出，而父进程并没有调用wait或waitpid获取子进程的状态信息，那么子进程的进程描述符仍然保存在系统中。这种进程称之为僵死进程。

任何一个子进程(init除外)在exit()之后，并非马上就消失掉，而是留下一个称为僵尸进程(Zombie)的数据结构，等待父进程处理

父进程一旦调用了wait就立即阻塞自己，由wait自动分析是否当前进程的某个子进程已经退出，如果让它找到了这样一个已经变成僵尸的子进程，wait就会收集这个子进程的信息，并把它彻底销毁后返回；如果没有找到这样一个子进程，wait就会一直阻塞在这里，直到有一个出现为止。当父进程忘了用wait()函数等待已终止的子进程时,子进程就会进入一种无父进程的状态,此时子进程就是僵尸进程。wait()要与fork()配套出现,如果在使用fork()之前调用wait(),wait()的返回值则为-1,正常情况下wait()的返回值为子进程的PID。如果先终止父进程,子进程将继续正常进行，只是它将由init进程(PID 1)继承,当子进程终止时,init进程捕获这个状态.


进程假死可以总结为不提供服务，但是还驻留在内存中


什么是linux系统假死
所谓假死，就是能ping通，但是ssh不上去；任何其它操作也都没反应，包括上面部署的nginx也打不开页面。

假死如何实现
有一个确定可以把系统搞成假死的办法是：主进程分配固定内存，然后不停的fork，并且在子进程里面sleep(100)。
也就是说，当主进程不停fork的时候，很快会把系统的物理内存用完，当物理内存不足时候，系统会开始使用swap；那么当swap不足时会触发oom killer进程；当oom killer杀掉了子进程，主进程会立刻fork新的子进程，并再次导致内存用完，再次触发oom killer进程，于是进入死循环。而且oom killer是系统底层优先级很高的内核线程，也在参与死循环。

系统假死为何能ping同无法连接
此时机器可以ping通，但是ssh无法上去。这是由于ping是在系统底层处理的，没有参与进程调度；sshd要参与进程调度，但是优先级没oom killer高，总得不到调度。



SIGINT、SIGTERM和SIGKILL区别


cat<<EOF| gofmt>aa.go
package utils

const (
    Version = "$version"
    Compile = "$compile"
)
EOF


git describe

date命令参数  date +"%F %T %z"

fork耗时

Fork新进程时，虽然可共享的数据内容不需要复制，但会复制之前进程空间的内存页表，这个复制是主线程来做的，会阻塞所有的读写操作，并且随着内存使用量越大耗时越长。


grep -E 'pattern1|pattern2'

grep -e pattern1  -e pattern2


systemctl 

systemctl show --property=Environment docker

