# Reactor模式

基于操作系统的多路复用机制实现，将事件的处理handler注册到事件分发器上，当对应的事件触发时调用响应的handler。所有的事件都可以由一个线程处理，也可以根据情况使用线程池来处理接收到的事件。

线程池本身可以缓解线程创建-销毁的代价，这样优化确实会好很多，不过还是存在一些问题的，就是线程的粒度太大。每一个线程把一次交互的事情全部做了，包括读取和返回，甚至连接，表面上似乎连接不在线程里，但是如果线程不够，有了新的连接，也无法得到处理，所以，目前的方案线程里可以看成要做三件事，连接，读取和写入。
线程同步的粒度太大了，限制了吞吐量。应该把一次连接的操作分为更细的粒度或者过程，这些更细的粒度是更小的线程。整个线程池的数目会翻倍，但是线程更简单，任务更加单一。这其实就是Reactor出现的原因，在Reactor中，这些被拆分的小线程或者子过程对应的是handler，每一种handler会出处理一种event。这里会有一个全局的管理者selector，我们需要把channel注册感兴趣的事件，那么这个selector就会不断在channel上检测是否有该类型的事件发生，如果没有，那么主线程就会被阻塞，否则就会调用相应的事件处理函数即handler来处理。典型的事件有连接，读取和写入，当然我们就需要为这些事件分别提供处理器，每一个处理器可以采用线程的方式实现。一个连接来了，显示被读取线程或者handler处理了，然后再执行写入，那么之前的读取就可以被后面的请求复用，吞吐量就提高了。