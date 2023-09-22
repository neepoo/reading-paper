# 是什么

Raft是管理日志复制的[共识算法](#共识算法是什么)。为了方便读者理解其算法，Raft将共识算法分为几个关键元素:

1. 领导选举(Leader Election)
2. 日志复制(Log Replication)
3. 安全性(Safety)
   并且它强制执行更高程度的一致性，以减少需要考虑的状态数量。Raft还包括一种用于更改集群成员的新机制，它使用重叠多数来确保安全性。

# 解决什么问题

用于解决分布式系统中的多种差若容忍问题。

# 共识算法是什么

原文是

```
Consensus algorithms allow a collection of machines
to work as a coherent group that can survive the fail-
ures of some of its members
```

翻译就是

```
共识算法允许一组机器作为一个协调一致的群体工作，即使其中的一些成员发生故障，仍能保持正常运作。
```

其中提到了一组机器，这好理解，分布式共识算法嘛。关键是`成员发生故障,整个系统如何正常运作`。后续会解决这个问题。

共识算法通常有如下属性:

1. 共识算法确保在所有非拜占庭条件下的安全性(始终不返回错误的结果).
    1. 非拜占庭条件是:假设分布式系统中的节点是可靠的。
2. 只要大多数节点是可用的(能够和节点及客户端通信)，共识算法就能正常运行。
    1. 五个节点的集群能够容错两个节点的故障。
    2. 假设服务器故障时会停止运行，但它们可以从稳定存储中恢复状态并重新加入集群。
3. 不依靠时间来保证日志的一致性。
4. 在常见情况下，只要集群中的大多数服务器对一轮远程过程调用作出响应，一个命令就可以完成；少数响应较慢的服务器不会影响整个系统的性能。

## Raft相较于其他共识算法最核心的特点

1. Raft使用更强的领导者形式，例如：日志(log entries)的流向只能从领导者流向其他节点。
2. 领导选举: Raft使用随机超时来选举领导者,别的共识算法大多是使用心跳(heartbeat).这高效且快速的解决了冲突。
3. 成员更改:Raft的用于更改集群中服务器集合的机制采用了一种新的联合共识方法，在过渡期间两个不同配置的多数派重叠。这使得集群在配置更改期间能够继续正常运行。

## 作者希望:

1. 我们还希望该算法能够促进系统构建者形成重要的直觉认知。重要的是，算法不仅需要能够工作，而且需要明显地解释其工作原理。

## 复制状态机 Replicated state machines

共识算法通常出现在复制状态机的背景下。在这种方法中，一组服务器上的状态机计算相同状态的相同副本，并且即使其中一些服务器宕机，它们仍然可以继续运行。
复制状态机通常使用复制日志来实现。保持复制日志的一致性是共识算法的目标。

服务器上的共识模块从客户端接收命令并将他们加到自己的日志中。他和其他服务器上的共识模块通信，以确保其他的服务器上的日志最终以相同的次序包含相同的请求(
或者说是命令)，
即便某些服务器发生故障。一旦日志被适当(properly)的复制到其他服务器,每个服务器的状态机就可以按照日志顺序处理他们。然后将结果返回客户端。从结果来看，
这些服务器看起来就像形成了一个单一且高度可靠的状态机。

## Raft共识算法

第一步是选举一个独一无二的领导者，随后领导者全权负责管理复制日志。

1. leader从客户端接收日志条目。
2. leader将日志条目复制到其他服务器上的日志中。
3. 告诉其余节点何时安全的将日志应用到他们的状态机中。

一个Raft集群包含多个服务器,在任意时刻,每个服务器处于以下三种状态之一: `leader`, `follower`, `candidate`.
在正常操作下只会有一个leader,follower是被动响应:他们不会主动发起请求，只会响应`leader`或者`candidate`的请求。
`leader`用于响应客户端的请求，如果客户端请求到了`follower`，`follower`会将请求转发给`leader`。
`candidate`用于选举`leader`。

Raft将时间划分为任意长度的任期(term)。任期是连续的整数.每次选举伴随着一个新的任期。如果一个candidate赢得了选举，那么它就会成为leader，然后开始为该任期服务。
如果这次选举没有选举出`leader`，那么就会进入下一轮选举，同时任期会增加。
每个服务器会存储当前的term,他是随着时间单调增的。每当服务器之间通信，彼此的term会交换.

1. 如果一个server的term比其他的小，那么它会更新自己的term为更大的值。
2. 如果candidate or leader发现自己的term已经过期了，他会将自己变成follower。
3. 如果一个server收到的请求的term比自己的小，那么他会拒绝这个请求。

领导者可能会失联，这种情况下会有新的领导者当选。

### 节点间通信

Raft servers间通过RPC通信,基本的共识算法需要两种RPC:

1. RequestVote RPC:election期间由candidate发起。
2. AppendEntries RPC:leader用于复制日志和提供心跳。

### 领导选举

Raft使用心跳机制来触发领导选举.集群启动是，所有节点状态都是followers,follower会继续保持follower状态，只要他们接收到来自candidate或者leader的RPC。
leader发送heartbeat(AppendEntries RPC但是不带有log entries)给所有的follower维持自己的权威.

如果一个follower在一段时间内(选举超时 election timeout)没有收到任何rpc,他会假设leader挂了，然后开始选举一个新的leader。
选举开始时，follower先增加自己的term,然后转换成candidate状态并开始选举过程。他会给自己投票并并行的向其他节点发送RequestVote
RPC。
除非以下三件其中之一发生，否则他会一直保持该状态。

1. 赢得选举
2. 另一个server成为了leader
3. 一段时间内没有新的leader产生

#### 赢得选举

他得到集群内统一任期的大多数节点的投票。每个节点在给定任期内只会给一个candidate投票。大多数规则保证了在给定term内，只会有一个candidate赢得选举。
一旦candidate赢得选举，他就会变成leader并开始发送心跳给其他节点建立权威，并且预防新的选举。

在candidate等待投票期间,candidate可能会收到自称为leader的AppendEntries RPC,如果leader的term>= candidate's current
term,candidate
承认leader的合法性并且转为follower状态。如果leader的term< candidate's current term,candidate拒绝这个请求并且继续保持candidate状态。

第三种可能的结果是一次选举没有选举出leader:如果在同一时刻大部分follower变成了candidate,会分票导致没有一个candidate赢得选举。这样每个candidate
就会开始增加自己的term,然后开始新一轮的选举。然而如果没有额外的机制保障，该过程可能会无限下去。

Raft使用随机选举超时来确保分票是罕见的，能够快速解决的。Raft选举超时通常是一个固定的间隙(150-300ms),
这样可以将服务器分散开来，以便在大多数情况下只有一个服务器会超时。每次选举先高一个随机选举超时，然后才进行真正的选举过程。

#### 领导选举时机？

### 日志复制

这里先把几个名词的概念给说清楚。`command`:leader接收到的客户端的请求，也是最终会作用到state machine上的东西,本质就是u8数组。
`entry`:存储在logs里的东西，除了`command`还有他在logs中的位置信息`index`以及他是在哪个任期`term`创建的.

#### 流程

`leader`接收客户端的请求`command`,将其封装成一个`entry`新增(append)到本地`logs`中,接着并行(parallel)地向所有的`follower`
发送
AppendEntries RPC让`follower`复制该`entry`到本地的`logs`中。当该`entry`被安全地(safely)复制后,`leader`
会apply该条`entry`到本地的`state machine`之后返回执行结果到客户端。无论什么原因(数据包丢失,`follower`响应慢等)只要`entry`
没有成功的复制，就会持续该过程直到成功.

`term`在`entry`中的作用是用来确定`logs`的不一致行为(`follower`和`leader`的日志不一样，如何让`follower`的变得和`leader`
一样)。

每当一条日志被提交(committed),提交的概念在这里的[第三条](#阅读中遇到的问题),该提交之前的所有`entries`
也会被提交，包括其他`leader`创建的`entry`。`leader`会追踪他已经提交的`entry`中的最大`index`记做`max_index`,每次AppendEntries
RPC会带着
`max_index`,这样其他的节点能够知道我需要将那些`entry`apply(按照log index 升序)到本地的`state machine`

#### 日志复制是如何保证log matching属性的

Log Matching Property说的是：

1. 如果两个不同日志中的条目具有相同的索引和任期，那么它们存储的是相同的命令。(If two entries in different logs have the
   same index
   and term, then they store the same command.)
2. 如果两个不同的日志中的条目具有相同的索引和任期，那么这意味着在这些条目之前的所有条目也是相同的。(If two entries in
   different logs have the same index
   and term, then the logs are identical in all preceding entries.)

第一条的保证来至在每个任期内,`leader`在给定`index`只会创建一个`entry`,并且该`entry`不会修改。
第二条的保证来自AppendEntries RPC的一致性检查:当发送AppendEntries RPC时，领导者会在其中包含其日志中紧随新条目之前的条目的索引和任期。
如果`follower`没有在`logs`找到该`entry`他会拒绝复制最新`entry`。通过归纳法:开始logs是空的，自然满足,然后AppendEntries
RPC的一致性检查又会在复制新`entry`时满足该属性。

作为结果，每当AppendEntries RPC返回成功，`leader`就知道`follower`的`logs`是和自己一模一样的。

#### `follower`和`leader`日志不一致如何解决。

节点之间的日志不一致是如何出现的?每当`leader`或者`follower`挂了，就会出现。这里可以去看论文的图7,这里我就不重复截图了。
不一致可能会是如下集中情况之一：

1. `follower`缺少`leader`中有的entry。
2. `follower`有`leader`中没有的`entry`
3. 上述两种情况都有.

Raft解决日志不一致的问题就是强制`follower`复制(duplicate)自己的`logs`。也就是说在`follower`中有冲突(conflicting)的`entry`
会被`leader`的`logs`覆盖。这里有一个问题，什么是`follower`中有冲突的日志了？论文中说的是通过AppendEntries
RPC间的信息交换，`leader`最终会知道该`leader`最新的和自己一致(index以及entry内的term一样)的`entry`,在这之后的`entry`
都是冲突的，需要被`leader`的`entries`覆盖的。

具体是怎么实现的了？`leader`会维护每个`leader`的一个属性叫做`nextIndex`，他的含义是`leader`将要发给`follower`下一个`entry`的`index`。
AppendEntries RPC会带着`nextIndex`过去，如果`follower`的日志和 `leader`的不一致，该次rpc会失败，接着`leader`会自减`nextIndex`的重复该过程，
直到节点之间的日志一致。在这次RPC内`follower`会删除不一致的`entry`同时复制`leader`的`entry`过来。这样的话日志就一致了。
这种场景下RPC可以优化，`follower`

### 安全性

如果任何节点应用了特定的日志条目到其状态机中，那么不会有其他节点在相同的日志索引位置应用不同的日志条目。
关键的安全属性还包括(无论什么时候，raft能保证如下属性):

1. election safety:给定任期(term)内，至多只会有一个leader当选.
2. leader append-only:leader只能通过追加日志条目的方式来更新日志，它不会覆盖或者删除日志中的条目。
3. Log Matching Property:如果两个日志中包含具有相同索引和任期的条目，则这些日志在给定索引之前的所有条目都是相同的。
4. Leader Completeness:如果在给定的任期中提交了一个日志条目，那么该条目将出现在所有更高任期的领导者的日志中。
    1. 这点在我看到日志复制时的理解是，leader如果已经将log entry apply到他的状态机中，它就会将结果返回给client.
       他会持续的发送appendEntries RPC给其他节点,直到其他节点存储了该entry。因为前提是非拜占庭。其他节点有了entry,就可以apply到他们的状态机中。也就是
       在别的节点上该entry也提交了。已经提交了的entry出现在更高term的leader的日志中我感觉就是显然的，这可能就是作者想表达的intuition。
5. State Machine Safety:如果一个服务器已经将一个给定索引的日志条目应用到其状态机中，那么其他服务器将永远不会为相同索引应用不同的日志条目。

## 阅读中遇到的问题

1. 成员发生故障,整个系统如何正常运作?
2. 什么叫做日志被适当的复制到其他服务器?是复制到集群中的大部分节点上吗？
3. 提交是什么意思？是指日志被复制了还是被应用到状态机了？
    1. 当leader将日志复制到大多数节点上，并且apply到自己的状态机后,该日志就被提交了。第二条说明了而非leader节点何时apply
       log entry.
    2. 一个`entry`被提交了，他之前的`entries`也是提交状态.
    3. The leader keeps track of the highest index it knows
       to be committed, and it includes that index in future
       AppendEntries RPCs (including heartbeats) so that the
       other servers eventually find out. Once a follower learns
       that a log entry is committed, it applies the entry to its
       local state machine (in log order).
4. safety property的3-5条，我不理解。
    1. 日志复制那里详细解释了log matching,现在还剩4-5.
5. 日志复制那里提到，只要appendEntries RPC失败会无限重复该步骤，那么如果follower宕机了，这不是多余的吗？希望在`safely`能得到答案。