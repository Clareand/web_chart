CREATE OR REPLACE FUNCTION public.f_check_session_refresh_token(ref_token character varying)
 RETURNS TABLE(is_not_expired boolean, customer_id uuid, remember_me boolean, id_session uuid, group_api jsonb)
 LANGUAGE plpgsql
AS $function$

BEGIN 

	return query 
	select
	case 
		when expired_at >= now() then true
		else false
	end as is_not_expired,
	s.customer_id,
	s.remember_me,
	s.id as id_session,
	s.group_api 
	from
		public.sessions s
	where
		(encode(public.hmac(id::text, key, 'sha256'), 'hex')) = REF_TOKEN;

END;

$function$
;


CREATE OR REPLACE FUNCTION public.f_check_user(user_name1 character varying)
 RETURNS TABLE(customer_id uuid, is_true boolean)
 LANGUAGE plpgsql
AS $function$
BEGIN 
	return query 
	select
		c.customer_id,
		case
			when COUNT(c.customer_id) > 0 then true
			else false
		end as is_true
	from
		public.customer c
	where (c.username = USER_NAME1)
	group by c.customer_id ;
END;
$function$
;

CREATE OR REPLACE FUNCTION public.f_create_session_refresh_token_user(user_id1 character varying, remember_me1 boolean, interval_month_remember1 character varying, interval_day_remember1 character varying, groupapi1 character varying)
 RETURNS TABLE(refresh_token text)
 LANGUAGE plpgsql
AS $function$
BEGIN 
	return query 
		with in_sessions as (
		insert
			into
				public.sessions (
					customer_id,
					remember_me,
					expired_at
				)
				select
					user_id1::uuid,
					REMEMBER_ME1,
					case
						when REMEMBER_ME1 = true then now()+ INTERVAL_MONTH_REMEMBER1::interval
						else now()+ INTERVAL_DAY_REMEMBER1::interval
					end
					returning *)
		select
			(encode(public.hmac(id::text, 'key', 'sha256'), 'hex')) as refresh_token
		from
			in_sessions;
END;
$function$
;


DROP FUNCTION f_get_log_data_user(character varying);
CREATE OR REPLACE FUNCTION public.f_get_log_data_user(user_id1 character varying)
 RETURNS TABLE(customer_id character varying, customer_email character varying, username character varying, created_at character varying, session uuid)
 LANGUAGE plpgsql
AS $function$
BEGIN 
	return query(
		select
			c.customer_id::varchar,
			c.customer_email,
			c.username,
			c.created_at::varchar,
			s.id 
		from
			public.customer c
		join
			public.sessions s on s.customer_id = c.customer_id 
		where
			c.customer_id = USER_ID1::uuid
			limit 1);
END;
$function$
;


CREATE OR REPLACE FUNCTION public.f_get_password_user(user_id1 character varying)
 RETURNS TABLE(credential character varying)
 LANGUAGE plpgsql
AS $function$
BEGIN 
	return query 
	select c.credential from public.customer c WHERE c.customer_id=user_id1
::uuid;
END;
$function$
;


CREATE OR REPLACE FUNCTION public.f_update_session(ref_token character varying, time_str character varying)
 RETURNS character varying
 LANGUAGE plpgsql
AS $function$
BEGIN 
	update public.sessions set expired_at=now()+ TIME_STR::interval
	where (encode(public.hmac(id::text, key, 'sha256'), 'hex')) = REF_TOKEN;
	return 'ok';
END;
$function$
;

CREATE OR REPLACE FUNCTION public.gen_salt(text)
 RETURNS text
 LANGUAGE c
 PARALLEL SAFE STRICT
AS '$libdir/pgcrypto', $function$pg_gen_salt$function$
;


CREATE OR REPLACE FUNCTION public.f_get_all_product()
	RETURNS table(product_id uuid, name character varying, stock int, price numeric)
	LANGUAGE plpgsql
AS $function$
	begin
		return query
		select 
			p.product_id,
			p."name",
			p.stock,
			p.price
		from public.products p 
		where p.is_deleted is false
		order by p."name" asc;

	END;
$function$
;
