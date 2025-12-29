--
-- PostgreSQL database dump
--

\restrict 6f7dJcerIXQhe3wg40YLgXMlkmRdY4ZY8Zk5Jf15BnT44EkuDtDfwPReg3qDAp0

-- Dumped from database version 18.1
-- Dumped by pg_dump version 18.1

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: scheduled_message_interval_unit; Type: TYPE; Schema: public; Owner: postgres
--

CREATE TYPE public.scheduled_message_interval_unit AS ENUM (
    'm',
    's'
);


ALTER TYPE public.scheduled_message_interval_unit OWNER TO postgres;

--
-- Name: scheduled_message_log_level; Type: TYPE; Schema: public; Owner: postgres
--

CREATE TYPE public.scheduled_message_log_level AS ENUM (
    'INFO',
    'WARN',
    'ERROR'
);


ALTER TYPE public.scheduled_message_log_level OWNER TO postgres;

--
-- Name: scheduled_message_schedule_type; Type: TYPE; Schema: public; Owner: postgres
--

CREATE TYPE public.scheduled_message_schedule_type AS ENUM (
    'interval',
    'cron'
);


ALTER TYPE public.scheduled_message_schedule_type OWNER TO postgres;

--
-- Name: scheduled_message_state; Type: TYPE; Schema: public; Owner: postgres
--

CREATE TYPE public.scheduled_message_state AS ENUM (
    'started',
    'stopped'
);


ALTER TYPE public.scheduled_message_state OWNER TO postgres;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: communication_platform; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.communication_platform (
    id integer NOT NULL,
    name character varying(64) NOT NULL,
    description text,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.communication_platform OWNER TO postgres;

--
-- Name: communication_platform_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.communication_platform_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.communication_platform_id_seq OWNER TO postgres;

--
-- Name: communication_platform_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.communication_platform_id_seq OWNED BY public.communication_platform.id;


--
-- Name: running_scheduled_message; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.running_scheduled_message (
    id integer NOT NULL,
    scheduled_message_id integer NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.running_scheduled_message OWNER TO postgres;

--
-- Name: running_scheduled_message_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.running_scheduled_message_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.running_scheduled_message_id_seq OWNER TO postgres;

--
-- Name: running_scheduled_message_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.running_scheduled_message_id_seq OWNED BY public.running_scheduled_message.id;


--
-- Name: scheduled_message; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.scheduled_message (
    id integer NOT NULL,
    name character varying(64) NOT NULL,
    user_defined_table_id integer NOT NULL,
    message text NOT NULL,
    rule text NOT NULL,
    schedule_type public.scheduled_message_schedule_type NOT NULL,
    communication_platform_id integer NOT NULL,
    description text,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.scheduled_message OWNER TO postgres;

--
-- Name: scheduled_message_cron; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.scheduled_message_cron (
    id integer NOT NULL,
    scheduled_message_id integer NOT NULL,
    minute character varying(8),
    hour character varying(8),
    day_of_month character varying(8),
    month character varying(8),
    day_of_week character varying(8),
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.scheduled_message_cron OWNER TO postgres;

--
-- Name: scheduled_message_cron_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.scheduled_message_cron_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.scheduled_message_cron_id_seq OWNER TO postgres;

--
-- Name: scheduled_message_cron_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.scheduled_message_cron_id_seq OWNED BY public.scheduled_message_cron.id;


--
-- Name: scheduled_message_execution_history; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.scheduled_message_execution_history (
    id integer NOT NULL,
    scheduled_message_id integer CONSTRAINT scheduled_message_execution_histo_scheduled_message_id_not_null NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.scheduled_message_execution_history OWNER TO postgres;

--
-- Name: scheduled_message_execution_history_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.scheduled_message_execution_history_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.scheduled_message_execution_history_id_seq OWNER TO postgres;

--
-- Name: scheduled_message_execution_history_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.scheduled_message_execution_history_id_seq OWNED BY public.scheduled_message_execution_history.id;


--
-- Name: scheduled_message_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.scheduled_message_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.scheduled_message_id_seq OWNER TO postgres;

--
-- Name: scheduled_message_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.scheduled_message_id_seq OWNED BY public.scheduled_message.id;


--
-- Name: scheduled_message_interval; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.scheduled_message_interval (
    id integer NOT NULL,
    scheduled_message_id integer NOT NULL,
    value integer NOT NULL,
    unit public.scheduled_message_interval_unit NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.scheduled_message_interval OWNER TO postgres;

--
-- Name: scheduled_message_interval_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.scheduled_message_interval_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.scheduled_message_interval_id_seq OWNER TO postgres;

--
-- Name: scheduled_message_interval_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.scheduled_message_interval_id_seq OWNED BY public.scheduled_message_interval.id;


--
-- Name: scheduled_message_log; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.scheduled_message_log (
    id integer NOT NULL,
    text text NOT NULL,
    level public.scheduled_message_log_level NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.scheduled_message_log OWNER TO postgres;

--
-- Name: scheduled_message_log_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.scheduled_message_log_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.scheduled_message_log_id_seq OWNER TO postgres;

--
-- Name: scheduled_message_log_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.scheduled_message_log_id_seq OWNED BY public.scheduled_message_log.id;


--
-- Name: scheduled_message_rule; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.scheduled_message_rule (
    id integer NOT NULL,
    name character varying(64) NOT NULL,
    description text,
    rule text,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.scheduled_message_rule OWNER TO postgres;

--
-- Name: scheduled_message_rule_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.scheduled_message_rule_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.scheduled_message_rule_id_seq OWNER TO postgres;

--
-- Name: scheduled_message_rule_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.scheduled_message_rule_id_seq OWNED BY public.scheduled_message_rule.id;


--
-- Name: scheduled_message_state_history; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.scheduled_message_state_history (
    id integer NOT NULL,
    state public.scheduled_message_state NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.scheduled_message_state_history OWNER TO postgres;

--
-- Name: scheduled_message_state_history_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.scheduled_message_state_history_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.scheduled_message_state_history_id_seq OWNER TO postgres;

--
-- Name: scheduled_message_state_history_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.scheduled_message_state_history_id_seq OWNED BY public.scheduled_message_state_history.id;


--
-- Name: user_defined_column; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.user_defined_column (
    id integer NOT NULL,
    user_defined_table_id integer NOT NULL,
    name character varying(64) NOT NULL,
    type character varying(64) NOT NULL,
    length integer,
    is_nullable boolean DEFAULT false NOT NULL,
    default_value text,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.user_defined_column OWNER TO postgres;

--
-- Name: user_defined_column_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.user_defined_column_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.user_defined_column_id_seq OWNER TO postgres;

--
-- Name: user_defined_column_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.user_defined_column_id_seq OWNED BY public.user_defined_column.id;


--
-- Name: user_defined_table; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.user_defined_table (
    id integer NOT NULL,
    name character varying(64) NOT NULL,
    description text,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.user_defined_table OWNER TO postgres;

--
-- Name: user_defined_table_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.user_defined_table_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.user_defined_table_id_seq OWNER TO postgres;

--
-- Name: user_defined_table_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.user_defined_table_id_seq OWNED BY public.user_defined_table.id;


--
-- Name: wazuh; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.wazuh (
    age integer DEFAULT 1
);


ALTER TABLE public.wazuh OWNER TO postgres;

--
-- Name: communication_platform id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.communication_platform ALTER COLUMN id SET DEFAULT nextval('public.communication_platform_id_seq'::regclass);


--
-- Name: running_scheduled_message id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.running_scheduled_message ALTER COLUMN id SET DEFAULT nextval('public.running_scheduled_message_id_seq'::regclass);


--
-- Name: scheduled_message id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.scheduled_message ALTER COLUMN id SET DEFAULT nextval('public.scheduled_message_id_seq'::regclass);


--
-- Name: scheduled_message_cron id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.scheduled_message_cron ALTER COLUMN id SET DEFAULT nextval('public.scheduled_message_cron_id_seq'::regclass);


--
-- Name: scheduled_message_execution_history id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.scheduled_message_execution_history ALTER COLUMN id SET DEFAULT nextval('public.scheduled_message_execution_history_id_seq'::regclass);


--
-- Name: scheduled_message_interval id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.scheduled_message_interval ALTER COLUMN id SET DEFAULT nextval('public.scheduled_message_interval_id_seq'::regclass);


--
-- Name: scheduled_message_log id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.scheduled_message_log ALTER COLUMN id SET DEFAULT nextval('public.scheduled_message_log_id_seq'::regclass);


--
-- Name: scheduled_message_rule id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.scheduled_message_rule ALTER COLUMN id SET DEFAULT nextval('public.scheduled_message_rule_id_seq'::regclass);


--
-- Name: scheduled_message_state_history id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.scheduled_message_state_history ALTER COLUMN id SET DEFAULT nextval('public.scheduled_message_state_history_id_seq'::regclass);


--
-- Name: user_defined_column id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_defined_column ALTER COLUMN id SET DEFAULT nextval('public.user_defined_column_id_seq'::regclass);


--
-- Name: user_defined_table id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_defined_table ALTER COLUMN id SET DEFAULT nextval('public.user_defined_table_id_seq'::regclass);


--
-- Data for Name: communication_platform; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.communication_platform (id, name, description, created_at) FROM stdin;
\.


--
-- Data for Name: running_scheduled_message; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.running_scheduled_message (id, scheduled_message_id, created_at) FROM stdin;
\.


--
-- Data for Name: scheduled_message; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.scheduled_message (id, name, user_defined_table_id, message, rule, schedule_type, communication_platform_id, description, created_at) FROM stdin;
\.


--
-- Data for Name: scheduled_message_cron; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.scheduled_message_cron (id, scheduled_message_id, minute, hour, day_of_month, month, day_of_week, created_at) FROM stdin;
\.


--
-- Data for Name: scheduled_message_execution_history; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.scheduled_message_execution_history (id, scheduled_message_id, created_at) FROM stdin;
\.


--
-- Data for Name: scheduled_message_interval; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.scheduled_message_interval (id, scheduled_message_id, value, unit, created_at) FROM stdin;
\.


--
-- Data for Name: scheduled_message_log; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.scheduled_message_log (id, text, level, created_at) FROM stdin;
\.


--
-- Data for Name: scheduled_message_rule; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.scheduled_message_rule (id, name, description, rule, created_at) FROM stdin;
3	new_rule	\N	test	2025-12-27 16:05:20.461726
\.


--
-- Data for Name: scheduled_message_state_history; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.scheduled_message_state_history (id, state, created_at) FROM stdin;
\.


--
-- Data for Name: user_defined_column; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.user_defined_column (id, user_defined_table_id, name, type, length, is_nullable, default_value, created_at) FROM stdin;
1	3	age	int4	255	f	1	2025-12-29 01:55:48.834818
\.


--
-- Data for Name: user_defined_table; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.user_defined_table (id, name, description, created_at) FROM stdin;
3	wazuh		2025-12-29 01:53:27.023857
\.


--
-- Data for Name: wazuh; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.wazuh (age) FROM stdin;
\.


--
-- Name: communication_platform_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.communication_platform_id_seq', 1, false);


--
-- Name: running_scheduled_message_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.running_scheduled_message_id_seq', 1, false);


--
-- Name: scheduled_message_cron_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.scheduled_message_cron_id_seq', 1, false);


--
-- Name: scheduled_message_execution_history_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.scheduled_message_execution_history_id_seq', 1, false);


--
-- Name: scheduled_message_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.scheduled_message_id_seq', 1, false);


--
-- Name: scheduled_message_interval_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.scheduled_message_interval_id_seq', 1, false);


--
-- Name: scheduled_message_log_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.scheduled_message_log_id_seq', 1, false);


--
-- Name: scheduled_message_rule_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.scheduled_message_rule_id_seq', 3, true);


--
-- Name: scheduled_message_state_history_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.scheduled_message_state_history_id_seq', 1, false);


--
-- Name: user_defined_column_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.user_defined_column_id_seq', 1, true);


--
-- Name: user_defined_table_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.user_defined_table_id_seq', 3, true);


--
-- Name: communication_platform communication_platform_name_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.communication_platform
    ADD CONSTRAINT communication_platform_name_key UNIQUE (name);


--
-- Name: communication_platform communication_platform_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.communication_platform
    ADD CONSTRAINT communication_platform_pkey PRIMARY KEY (id);


--
-- Name: running_scheduled_message running_scheduled_message_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.running_scheduled_message
    ADD CONSTRAINT running_scheduled_message_pkey PRIMARY KEY (id);


--
-- Name: running_scheduled_message running_scheduled_message_scheduled_message_id_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.running_scheduled_message
    ADD CONSTRAINT running_scheduled_message_scheduled_message_id_key UNIQUE (scheduled_message_id);


--
-- Name: scheduled_message_cron scheduled_message_cron_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.scheduled_message_cron
    ADD CONSTRAINT scheduled_message_cron_pkey PRIMARY KEY (id);


--
-- Name: scheduled_message_execution_history scheduled_message_execution_history_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.scheduled_message_execution_history
    ADD CONSTRAINT scheduled_message_execution_history_pkey PRIMARY KEY (id);


--
-- Name: scheduled_message_interval scheduled_message_interval_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.scheduled_message_interval
    ADD CONSTRAINT scheduled_message_interval_pkey PRIMARY KEY (id);


--
-- Name: scheduled_message_log scheduled_message_log_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.scheduled_message_log
    ADD CONSTRAINT scheduled_message_log_pkey PRIMARY KEY (id);


--
-- Name: scheduled_message scheduled_message_name_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.scheduled_message
    ADD CONSTRAINT scheduled_message_name_key UNIQUE (name);


--
-- Name: scheduled_message scheduled_message_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.scheduled_message
    ADD CONSTRAINT scheduled_message_pkey PRIMARY KEY (id);


--
-- Name: scheduled_message_rule scheduled_message_rule_name_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.scheduled_message_rule
    ADD CONSTRAINT scheduled_message_rule_name_key UNIQUE (name);


--
-- Name: scheduled_message_rule scheduled_message_rule_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.scheduled_message_rule
    ADD CONSTRAINT scheduled_message_rule_pkey PRIMARY KEY (id);


--
-- Name: scheduled_message_state_history scheduled_message_state_history_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.scheduled_message_state_history
    ADD CONSTRAINT scheduled_message_state_history_pkey PRIMARY KEY (id);


--
-- Name: user_defined_column user_defined_column_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_defined_column
    ADD CONSTRAINT user_defined_column_pkey PRIMARY KEY (id);


--
-- Name: user_defined_table user_defined_table_name_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_defined_table
    ADD CONSTRAINT user_defined_table_name_key UNIQUE (name);


--
-- Name: user_defined_table user_defined_table_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_defined_table
    ADD CONSTRAINT user_defined_table_pkey PRIMARY KEY (id);


--
-- Name: running_scheduled_message running_scheduled_message_scheduled_message_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.running_scheduled_message
    ADD CONSTRAINT running_scheduled_message_scheduled_message_id_fkey FOREIGN KEY (scheduled_message_id) REFERENCES public.scheduled_message(id) ON DELETE CASCADE;


--
-- Name: scheduled_message scheduled_message_communication_platform_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.scheduled_message
    ADD CONSTRAINT scheduled_message_communication_platform_id_fkey FOREIGN KEY (communication_platform_id) REFERENCES public.communication_platform(id) ON DELETE CASCADE;


--
-- Name: scheduled_message_cron scheduled_message_cron_scheduled_message_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.scheduled_message_cron
    ADD CONSTRAINT scheduled_message_cron_scheduled_message_id_fkey FOREIGN KEY (scheduled_message_id) REFERENCES public.scheduled_message(id) ON DELETE CASCADE;


--
-- Name: scheduled_message_execution_history scheduled_message_execution_history_scheduled_message_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.scheduled_message_execution_history
    ADD CONSTRAINT scheduled_message_execution_history_scheduled_message_id_fkey FOREIGN KEY (scheduled_message_id) REFERENCES public.scheduled_message(id) ON DELETE CASCADE;


--
-- Name: scheduled_message_interval scheduled_message_interval_scheduled_message_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.scheduled_message_interval
    ADD CONSTRAINT scheduled_message_interval_scheduled_message_id_fkey FOREIGN KEY (scheduled_message_id) REFERENCES public.scheduled_message(id) ON DELETE CASCADE;


--
-- Name: scheduled_message scheduled_message_user_defined_table_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.scheduled_message
    ADD CONSTRAINT scheduled_message_user_defined_table_id_fkey FOREIGN KEY (user_defined_table_id) REFERENCES public.user_defined_table(id) ON DELETE CASCADE;


--
-- Name: user_defined_column user_defined_column_user_defined_table_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_defined_column
    ADD CONSTRAINT user_defined_column_user_defined_table_id_fkey FOREIGN KEY (user_defined_table_id) REFERENCES public.user_defined_table(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

\unrestrict 6f7dJcerIXQhe3wg40YLgXMlkmRdY4ZY8Zk5Jf15BnT44EkuDtDfwPReg3qDAp0

