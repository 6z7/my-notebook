通过索引，减少数据量，磁盘每个block能否存储更多的数据，总的数据所需block更好，磁盘io更少。

指向索引的索引由于只有被指向的索引的指针，数据量更小，这种二级索引被成为稀疏索引

二叉搜索树为什么不用二叉树存储索引呢？

一次磁盘IO加载的数据有限，二叉树瘦高，一次加载包含有效分支和无效分支，所以造成有效分支的数据量少，如果目标值靠后则需要更多的IO来加载更多的数据。


结构+规则 构成满足条件的数据结构


B+数矮胖，每个节点存储更多的数据，高度越低，最终到达叶子节点的路径就越短，即所需的磁盘IO就越少