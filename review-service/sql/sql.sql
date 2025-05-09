

create schema comment-service;
create table if not exists review_appeal_info
(
    id         bigint unsigned auto_increment comment '主键'
    primary key,
    create_by  varchar(48)   default ''                not null comment '创建方标识',
    update_by  varchar(48)   default ''                not null comment '更新方标识',
    create_at  timestamp     default CURRENT_TIMESTAMP not null comment '创建时间',
    update_at  timestamp     default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP comment '更新时间',
    delete_at  timestamp                               null comment '逻辑删除标记',
    version    int unsigned  default '0'               not null comment '乐观锁标记',
    appeal_id  bigint        default 0                 not null comment '回复id',
    review_id  bigint        default 0                 not null comment '评价id',
    store_id   bigint        default 0                 not null comment '店铺id',
    status     tinyint       default 10                not null comment '状态:10待审核；20申诉通过；30申诉驳回',
    reason     varchar(255)                            not null comment '申诉原因类别',
    content    varchar(255)                            not null comment '申诉内容描述',
    pic_info   varchar(1024) default ''                not null comment '媒体信息：图片',
    video_info varchar(1024) default ''                not null comment '媒体信息：视频',
    op_remarks varchar(512)  default ''                not null comment '运营备注',
    op_user    varchar(64)   default ''                not null comment '运营者标识',
    ext_json   varchar(1024) default ''                not null comment '信息扩展',
    ctrl_json  varchar(1024) default ''                not null comment '控制扩展'
    )
    comment '评价商家申诉表' engine = InnoDB
    charset = utf8mb4;

create index idx_appeal_id
    on review_appeal_info (appeal_id)
    comment '申诉id索引';

-- comment on index idx_appeal_id not supported: 申诉id索引

create index idx_delete_at
    on review_appeal_info (delete_at)
    comment '逻辑删除索引';

-- comment on index idx_delete_at not supported: 逻辑删除索引

create index idx_review_id
    on review_appeal_info (review_id)
    comment '评价id索引';

-- comment on index idx_review_id not supported: 评价id索引

create index idx_store_id
    on review_appeal_info (store_id)
    comment '店铺id索引';

-- comment on index idx_store_id not supported: 店铺id索引

create table if not exists review_info
(
    id              bigint unsigned auto_increment comment '主键'
    primary key,
    create_by       varchar(48)   default ''                not null comment '创建方标识',
    update_by       varchar(48)   default ''                not null comment '更新方标识',
    create_at       timestamp     default CURRENT_TIMESTAMP not null comment '创建时间',
    update_at       timestamp     default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP comment '更新时间',
    delete_at       timestamp                               null comment '逻辑删除标记',
    version         int unsigned  default '0'               not null comment '乐观锁标记',
    review_id       bigint        default 0                 not null comment '评价id',
    content         varchar(512)                            not null comment '评价内容',
    score           tinyint       default 0                 not null comment '评分',
    service_score   tinyint       default 0                 not null comment '商家服务评分',
    express_score   tinyint       default 0                 not null comment '物流评分',
    has_media       tinyint       default 0                 not null comment '是否有图或视频',
    order_id        bigint        default 0                 not null comment '订单id',
    sku_id          bigint        default 0                 not null comment 'sku id',
    spu_id          bigint        default 0                 not null comment 'spu id',
    store_id        bigint        default 0                 not null comment '店铺id',
    user_id         bigint        default 0                 not null comment '用户id',
    anonymous       tinyint       default 0                 not null comment '是否匿名',
    tags            varchar(1024) default ''                not null comment '标签json',
    pic_info        varchar(1024) default ''                not null comment '媒体信息：图片',
    video_info      varchar(1024) default ''                not null comment '媒体信息：视频',
    status          tinyint       default 10                not null comment '状态:10待审核；20审核通过；30审核不通过；40隐藏',
    is_default      tinyint       default 0                 not null comment '是否默认评价',
    has_reply       tinyint       default 0                 not null comment '是否有商家回复:0无;1有',
    op_reason       varchar(512)  default ''                not null comment '运营审核拒绝原因',
    op_remarks      varchar(512)  default ''                not null comment '运营备注',
    op_user         varchar(64)   default ''                not null comment '运营者标识',
    goods_snapshoot varchar(2048) default ''                not null comment '商品快照信息',
    ext_json        varchar(1024) default ''                not null comment '信息扩展',
    ctrl_json       varchar(1024) default ''                not null comment '控制扩展'
    )
    comment '评价表' engine = InnoDB
    charset = utf8mb4;

create index idx_delete_at
    on review_info (delete_at)
    comment '逻辑删除索引';

-- comment on index idx_delete_at not supported: 逻辑删除索引

create index idx_order_id
    on review_info (order_id)
    comment '订单id索引';

-- comment on index idx_order_id not supported: 订单id索引

create index idx_review_id
    on review_info (review_id)
    comment '评价id索引';

-- comment on index idx_review_id not supported: 评价id索引

create index idx_user_id
    on review_info (user_id)
    comment '用户id索引';

-- comment on index idx_user_id not supported: 用户id索引

create table if not exists review_reply_info
(
    id         bigint unsigned auto_increment comment '主键'
    primary key,
    create_by  varchar(48)   default ''                not null comment '创建方标识',
    update_by  varchar(48)   default ''                not null comment '更新方标识',
    create_at  timestamp     default CURRENT_TIMESTAMP not null comment '创建时间',
    update_at  timestamp     default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP comment '更新时间',
    delete_at  timestamp                               null comment '逻辑删除标记',
    version    int unsigned  default '0'               not null comment '乐观锁标记',
    reply_id   bigint        default 0                 not null comment '回复id',
    review_id  bigint        default 0                 not null comment '评价id',
    store_id   bigint        default 0                 not null comment '店铺id',
    content    varchar(512)                            not null comment '评价内容',
    pic_info   varchar(1024) default ''                not null comment '媒体信息：图片',
    video_info varchar(1024) default ''                not null comment '媒体信息：视频',
    ext_json   varchar(1024) default ''                not null comment '信息扩展',
    ctrl_json  varchar(1024) default ''                not null comment '控制扩展'
    )
    comment '评价商家回复表' engine = InnoDB
    charset = utf8mb4;

create index idx_delete_at
    on review_reply_info (delete_at)
    comment '逻辑删除索引';

-- comment on index idx_delete_at not supported: 逻辑删除索引

create index idx_reply_id
    on review_reply_info (reply_id)
    comment '回复id索引';

-- comment on index idx_reply_id not supported: 回复id索引

create index idx_review_id
    on review_reply_info (review_id)
    comment '评价id索引';

-- comment on index idx_review_id not supported: 评价id索引

create index idx_store_id
    on review_reply_info (store_id)
    comment '店铺id索引';

-- comment on index idx_store_id not supported: 店铺id索引

