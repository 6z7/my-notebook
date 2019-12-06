先写一个简单的chan例子,查看生成的汇编代码可以发现关键在**runtime.chanrecv1**与**runtime.chansend1**，下面我们来具体分析。
     
        func main() {
            ch := make(chan int)
            go func() {
                ch<-1
            }()
            fmt.Println(<-ch)
        }


        0x006f 00111 (demo2.go:10)	MOVQ	AX, (SP)
        0x0073 00115 (demo2.go:10)	PCDATA	$0, $1
        0x0073 00115 (demo2.go:10)	LEAQ	""..autotmp_3+48(SP), AX
        0x0078 00120 (demo2.go:10)	PCDATA	$0, $0
        0x0078 00120 (demo2.go:10)	MOVQ	AX, 8(SP)
        //接收消息
        0x007d 00125 (demo2.go:10)	CALL	runtime.chanrecv1(SB)
        ....
        0x00df 00223 (demo2.go:10)	MOVQ	AX, (SP)
        0x00e3 00227 (demo2.go:10)	MOVQ	$1, 8(SP)
        0x00ec 00236 (demo2.go:10)	MOVQ	$1, 16(SP)
        0x00f5 00245 (demo2.go:10)	CALL	fmt.Println(SB)
        0x00fa 00250 (demo2.go:11)	MOVQ	120(SP), BP
        0x00ff 00255 (demo2.go:11)	SUBQ	$-128, SP
        0x0103 00259 (demo2.go:11)	RET
        0x0104 00260 (demo2.go:11)	NOP
    "".main.func1 STEXT size=72 args=0x8 locals=0x18
        0x0000 00000 (demo2.go:7)	TEXT	"".main.func1(SB), ABIInternal, $24-8
        0x0026 00038 (demo2.go:8)	LEAQ	""..stmp_0(SB), AX
        0x002d 00045 (demo2.go:8)	PCDATA	$0, $0
        0x002d 00045 (demo2.go:8)	MOVQ	AX, 8(SP)
        //写入消息
        0x0032 00050 (demo2.go:8)	CALL	runtime.chansend1(SB)
        0x0037 00055 (demo2.go:9)	MOVQ	16(SP), BP
        0x003c 00060 (demo2.go:9)	ADDQ	$24, SP
        0x0040 00064 (demo2.go:9)	RET
        0x0041 00065 (demo2.go:9)	NOP

## chan在运行时的结构

    //chan结构
    type hchan struct {
        //当前缓冲中的消息数量
        qcount   uint           // total data in the queue
        //chan缓冲大小 make(chan int,10)
        dataqsiz uint           // size of the circular queue
        //chan缓冲数组
        buf      unsafe.Pointer // points to an array of dataqsiz elements
        //chn中元素类型的大小 make(chan int)
        elemsize uint16
        //chan是否关闭
        closed   uint32
        //chn中元素类型
        elemtype *_type // element type
        //发送数据保存缓冲位置
        sendx    uint   // send index
        //接收索引
        recvx    uint   // receive index
        //chnn接收者等待队列
        recvq    waitq  // list of recv waiters
        //chann发送者等待队列
        sendq    waitq  // list of send waiters
        //互斥锁       
        lock mutex
    }

## runtime.chanrecv1 ##

    // 接收消息保存到ep
    //ep:保存接收的数据,为nil的话将忽略接收数据
    // 如果block=false没有数据到达时 returns (false, false)

    func chanrecv(c *hchan, ep unsafe.Pointer, block bool) (selected, received bool) {

        if debugChan {
            print("chanrecv: chan=", c, "\n")
        }

        // 如果在 nil channel 上进行 recv 操作，会永远阻塞
        if c == nil {
            if !block {
                return
            }
            gopark(nil, nil, waitReasonChanReceiveNilChan, traceEvGoStop, 2)
            throw("unreachable")
        }
        
        //非阻塞&&chan未关闭&&(没有缓冲&&没有发送者阻塞||有缓冲&&缓冲为空)
        //满足以上条件直接返回
        if !block&& (c.dataqsiz == 0 && c.sendq.first == nil ||
            c.dataqsiz > 0 && atomic.Loaduint(&c.qcount) == 0) &&
            atomic.Load(&c.closed) == 0 {
            return
        }

        var t0 int64
        if blockprofilerate > 0 {
            t0 = cputicks()
        }

        lock(&c.lock)

        //chan已经关闭且缓冲中没有消息可读,直接返回
        //通过此处可以看到从已经关闭的chan读取数据时不会报错
        if c.closed != 0 && c.qcount == 0 {           
            unlock(&c.lock)
            if ep != nil {
                typedmemclr(c.elemtype, ep)
            }
            return true, false
        }

        //有发送者被挂起,直接将发送者的消息返回给接收者
        if sg := c.sendq.dequeue(); sg != nil {	 
            recv(c, sg, ep, func() { unlock(&c.lock) }, 3)
            return true, true
        }

        //缓冲中有消息
        if c.qcount > 0 {
            // Receive directly from queue
            //缓冲中c.recvx位置处的消息
            qp := chanbuf(c, c.recvx)           
            if ep != nil {
                //将缓冲中的数据复制到接收者ep
                typedmemmove(c.elemtype, ep, qp)
            }
            //清空缓冲中c.recvx位置处的值,留给发送者使用
            typedmemclr(c.elemtype, qp)
            //接收索引+1，下次读取缓冲中的下一个值
            c.recvx++
            //缓冲中的数据已经接收完毕(索引从0开始)从头开始
            if c.recvx == c.dataqsiz {
                c.recvx = 0
            }
            //消息数量-1
            c.qcount--
            unlock(&c.lock)
            return true, true
        }

        //非阻塞时如果没有收到消息直接返回
        if !block {
            unlock(&c.lock)
            return false, false
        }

        //没有数据可以接收，则挂起当前g       
        gp := getg()
        //获取Sudog保存当前信息
        mysg := acquireSudog()
        mysg.releasetime = 0
        if t0 != 0 {
            mysg.releasetime = -1
        }       
        //保存接收者,唤醒时会将发送的消息赋值给elm
        mysg.elem = ep
        mysg.waitlink = nil
        gp.waiting = mysg
        mysg.g = gp
        mysg.isSelect = false
        mysg.c = c
        gp.param = nil
        c.recvq.enqueue(mysg)
        goparkunlock(&c.lock, waitReasonChanReceive, traceEvGoBlockRecv, 3)

        //g被唤醒后续操作      
        if mysg != gp.waiting {
            throw("G waiting list is corrupted")
        }
        gp.waiting = nil
        if mysg.releasetime > 0 {
            blockevent(mysg.releasetime-t0, 2)
        }
        //被唤醒后 如果gp.param == nil则代表chan已经被关闭
        //close时会将阻塞的接收者的gp.param设置为nil,发送者唤醒时会将gp.param赋值为被唤醒者的sudog
        closed := gp.param == nil
        gp.param = nil
        mysg.c = nil
        releaseSudog(mysg)
        return true, !closed
    }


    // 从挂起的发送者上接收消息,此时缓冲中的消息数量不会变

    // c:接受者的chan
    // sg:代表被阻塞的发送者
    // ep:保存数据的接收者
    // unlockf:唤醒g时的回调操作
    // 如果没有缓冲即同步chan,直接从发送者copy数据到ep,之后唤醒被阻塞的g
    // 如果是异步chan,则先chan缓冲中copy数据到ep,再将发送者的数据放到缓冲中,之后唤醒被阻塞的g     
    func recv(c *hchan, sg *sudog, ep unsafe.Pointer, unlockf func(), skip int) {
        //没有设置缓冲 make(chan int)这种形式
        if c.dataqsiz == 0 {           
            if ep != nil {                
                //将发送数据直接copy到接收者上
                recvDirect(c.elemtype, sg, ep)
            }
        } else {            
            // chan有缓冲 只有缓冲已经填满，才会走到这里(只有带缓冲的chan已经满才会造成发送者阻塞才会走到此处)
            //缓冲中的c.recvx处保存的数据
            qp := chanbuf(c, c.recvx) 
            //从缓冲中copy数据到接收者ep
            if ep != nil {
                typedmemmove(c.elemtype, ep, qp)
            }           
            //qp位置处的值已经被消费 复制给了ep
            //从sender中copy数据到缓冲中
            //从缓冲队列c.recvx取出值,并将发送者的数据保存到队列中,此时缓冲数量没有变换
            typedmemmove(c.elemtype, qp, sg.elem)
            //消费者位置+1
            c.recvx++
            //缓冲已满
            if c.recvx == c.dataqsiz {
                c.recvx = 0
            }
            //缓冲已满时发送者会将c.sendx = 0,即此处c.sendx=0
		    //c.recvx是缓冲中下次被访问的位置，空出来后可以给发送者使用，
		    //所以将发送者的位置设置为下次空出来的位置			    	    
            c.sendx = c.recvx // c.sendx = (c.sendx+1) % c.dataqsiz
        }
        sg.elem = nil
        gp := sg.g
        unlockf()
        //指向阻塞的发送者 gp.param==nil代表chan已经被关闭
        gp.param = unsafe.Pointer(sg)
        if sg.releasetime != 0 {
            sg.releasetime = cputicks()
        }
        //唤醒发送者g
        goready(gp, skip+1)
    }