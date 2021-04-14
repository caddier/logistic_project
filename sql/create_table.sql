create database logist;
use logist;
drop table if exists register_user;
create table register_user(
	id bigint auto_increment primary key,
    user_wechat_id varchar(64) default '' comment 'user wechat id ',
    phone_no varchar(32) default '',
    create_time datetime default now()
);

drop table if exists logistic_thirdparty;
create table logistic_thirdparty(
	id bigint auto_increment primary key,
    third_party_name varchar(64),
    third_party_api varchar(1024) default '',
	third_parth_table_name varchar(32) not null,
    create_time datetime default now()
);


drop table  if exists logistic_order;
create table logistic_order(
	id bigint auto_increment primary key,
    user_id bigint default 0,
    our_order_id varchar(128) not null,
    third_party_order_id varchar(128) not null,
    third_party_id bigint not null,
    batchsn varchar(64),
    expressno varchar(64),
    expressname varchar(64),
    recipient varchar(64) ,
    recipienttel varchar(64) ,
    recipientaddress varchar(512),
    recipientcardno varchar(128) comment 'recipient id',
    recipientshopinfo varchar(128) comment 'item name',
    last_status varchar(256) comment 'save last status info ',
    statustime varchar(64),
    is_finished varchar(1) default 'N' comment 'Y / N ',
    create_time datetime default now()
);

create UNIQUE index idx_our_order_id on logistic_order(our_order_id);
create UNIQUE index idx_3rd_order_id on logistic_order(third_party_order_id);

drop table if exists logistic_order_status;
create table logistic_order_status(
	id bigint auto_increment primary key,
    third_party_order_id varchar(128) not null,
    order_status_desc varchar(512) default '',
    order_status_time datetime,
    create_time datetime default now()
);