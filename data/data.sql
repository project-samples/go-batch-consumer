create table users (
	id varchar(40) not null,
	username varchar(80),
	email varchar(100),
	url varchar(100),
	phone varchar(20),
	active char(1),
	locked int,
	date_of_birth timestamp,
	primary key (id)
);