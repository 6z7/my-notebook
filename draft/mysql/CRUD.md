## 新增

```sql
insert into room(id,userid,name) values (1,2,'');
insert into room  values (5,2,''),(3,3,'')

insert into room select 6,6,'';
insert into room select 9 as a,9 as b,'' as c;
```

## 删除

```sql
delete from room where id=1;
```

## 修改

```sql
update room set name='aa' where id=9 or id =1;
update room set name='aa' where id in (5,6);
```

## 查询

```sql
select a.dept_no, a.dept_name, b.from_date
from departments a
         inner join dept_emp b on a.dept_no = b.dept_no
where a.dept_no = 'd005'

select a.dept_no, a.dept_name, b.from_date
from departments a
         left join dept_emp b on a.dept_no = b.dept_no
where a.dept_no = 'd005';

#在模糊查询中，%表示任意0个或多个字符；_表示任意单个字符（有且仅有）
select * from aa  where name like '_e%';


# 从0开始，取2个
# MySQL并不是跳过offset行，而是取offset+N行，然后返回放弃前offset行，返回N行，
# 当offset特别大，然后单条数据也很大的时候，每次查询需要获取的数据就越多就会很慢。
select * from employees.dept_emp limit 1,1;
select * from employees.dept_emp limit 1 offset 1;

```