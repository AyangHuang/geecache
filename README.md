1. 学习geetoto的geecache

2. 改动：

   1. 增加algorithm.Cache接口，利用工厂模式创建缓存算法实例。cache 依赖LRU改成依赖algorithm.Cache接口，依赖倒转实现解耦。可拓展多种算法。
   2. 增加LFU缓存算法。




![](./image/流程.png)