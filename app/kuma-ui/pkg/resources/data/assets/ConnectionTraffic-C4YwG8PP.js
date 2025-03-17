import{_ as T,r as C,o as d,m as x,w as g,s,a as v,e as r,d as M,k as j,q as y,t as n,p as a,c,F as m,v as q,n as z,b as K}from"./index-Bi3CXAeE.js";import{T as W}from"./TagList-DYjb-hhI.js";const H=["B","kB","MB","GB","TB","PB","EB","ZB","YB"],J=["B","KiB","MiB","GiB","TiB","PiB","EiB","ZiB","YiB"],Q=["b","kbit","Mbit","Gbit","Tbit","Pbit","Ebit","Zbit","Ybit"],R=["b","kibit","Mibit","Gibit","Tibit","Pibit","Eibit","Zibit","Yibit"],Z=(o,i,e)=>{let u=o;return typeof i=="string"||Array.isArray(i)?u=o.toLocaleString(i,e):(i===!0||e!==void 0)&&(u=o.toLocaleString(void 0,e)),u};function k(o,i){if(!Number.isFinite(o))throw new TypeError(`Expected a finite number, got ${typeof o}: ${o}`);i={bits:!1,binary:!1,space:!0,...i};const e=i.bits?i.binary?R:Q:i.binary?J:H,u=i.space?" ":"";if(i.signed&&o===0)return` 0${u}${e[0]}`;const p=o<0,t=p?"-":i.signed?"+":"";p&&(o=-o);let f;if(i.minimumFractionDigits!==void 0&&(f={minimumFractionDigits:i.minimumFractionDigits}),i.maximumFractionDigits!==void 0&&(f={maximumFractionDigits:i.maximumFractionDigits,...f}),o<1){const B=Z(o,i.locale,f);return t+B+u+e[0]}const _=Math.min(Math.floor(i.binary?Math.log(o)/Math.log(1024):Math.log10(o)/3),e.length-1);o/=(i.binary?1024:1e3)**_,f||(o=o.toPrecision(3));const $=Z(Number(o),i.locale,f),b=e[_];return t+$+u+b}const h={},tt={class:"card"},et={class:"title"},it={class:"body"};function st(o,i){const e=C("XCard");return d(),x(e,{class:"data-card"},{default:g(()=>[s("dl",null,[s("div",tt,[s("dt",et,[v(o.$slots,"title",{},void 0,!0)]),i[0]||(i[0]=r()),s("dd",it,[v(o.$slots,"default",{},void 0,!0)])])])]),_:3})}const A=T(h,[["render",st],["__scopeId","data-v-719ec237"]]),rt={class:"title"},ot={key:0},nt={key:1},at={"data-testid":"grpc-success"},dt={"data-testid":"grpc-failure"},lt={"data-testid":"rq-2xx"},ct={"data-testid":"rq-4xx"},ft={"data-testid":"rq-5xx"},ut={"data-testid":"connections-total"},pt={key:0,"data-testid":"bytes-received"},_t={key:1,"data-testid":"bytes-sent"},mt=M({__name:"ConnectionCard",props:{protocol:{},service:{default:""},traffic:{default:void 0},direction:{default:"downstream"},portName:{default:void 0}},setup(o){const{t:i}=j(),e=o,u=p=>{const t=p.target;if(p.isTrusted&&t.nodeName.toLowerCase()!=="a"){const f=t.closest(".service-traffic-card, a");if(f){const _=f.nodeName.toLowerCase()==="a"?f:f.querySelector("[data-action]");_!==null&&"click"in _&&typeof _.click=="function"&&_.click()}}};return(p,t)=>{const f=C("XBadge"),_=C("XProgress");return d(),x(A,{class:"service-traffic-card",onClick:u},{title:g(()=>[e.service.length>0?(d(),x(W,{key:0,tags:[{label:"kuma.io/service",value:e.service}]},null,8,["tags"])):y("",!0),t[1]||(t[1]=r()),s("div",rt,[e.protocol!==""?(d(),x(f,{key:0,class:"protocol",appearance:e.protocol==="passthrough"?"success":"info"},{default:g(()=>[r(n(a(i)(`data-planes.components.service_traffic_card.protocol.${e.protocol}`,{},{defaultMessage:a(i)(`http.api.value.${e.protocol}`)})),1)]),_:1},8,["appearance"])):y("",!0),t[0]||(t[0]=r()),v(p.$slots,"default",{},void 0,!0)])]),default:g(()=>{var $,b,B,w,E,F,D,X,L,P,G,Y;return[t[24]||(t[24]=r()),e.portName?(d(),c("dl",ot,[s("div",null,[t[2]||(t[2]=s("dt",null,"Name",-1)),t[3]||(t[3]=r()),s("dd",null,n(e.portName),1)])])):y("",!0),t[25]||(t[25]=r()),e.traffic?(d(),c("dl",nt,[e.protocol==="passthrough"?(d(!0),c(m,{key:0},q([["http","tcp"].reduce((l,N)=>{var V;const U=e.direction;return Object.entries(((V=e.traffic)==null?void 0:V[N])||{}).reduce((I,[S,O])=>[`${U}_cx_tx_bytes_total`,`${U}_cx_rx_bytes_total`].includes(S)?{...I,[S]:O+(I[S]??0)}:I,l)},{})],(l,N)=>(d(),c(m,{key:N},[s("div",null,[s("dt",null,n(a(i)("data-planes.components.service_traffic_card.tx")),1),t[4]||(t[4]=r()),s("dd",null,n(a(k)(l.downstream_cx_rx_bytes_total??0)),1)]),t[6]||(t[6]=r()),s("div",null,[s("dt",null,n(a(i)("data-planes.components.service_traffic_card.rx")),1),t[5]||(t[5]=r()),s("dd",null,n(a(k)(l.downstream_cx_tx_bytes_total??0)),1)])],64))),128)):e.protocol==="grpc"?(d(),c(m,{key:1},[s("div",at,[s("dt",null,n(a(i)("data-planes.components.service_traffic_card.grpc_success")),1),t[7]||(t[7]=r()),s("dd",null,n(a(i)("common.formats.integer",{value:($=e.traffic.grpc)==null?void 0:$.success})),1)]),t[9]||(t[9]=r()),s("div",dt,[s("dt",null,n(a(i)("data-planes.components.service_traffic_card.grpc_failure")),1),t[8]||(t[8]=r()),s("dd",null,n(a(i)("common.formats.integer",{value:(b=e.traffic.grpc)==null?void 0:b.failure})),1)])],64)):e.protocol.startsWith("http")?(d(),c(m,{key:2},[(d(!0),c(m,null,q([((B=e.traffic.http)==null?void 0:B[`${e.direction}_rq_1xx`])??0].filter(l=>l!==0),l=>(d(),c("div",{key:l,"data-testid":"rq-1xx"},[s("dt",null,n(a(i)("data-planes.components.service_traffic_card.1xx")),1),t[10]||(t[10]=r()),s("dd",null,n(a(i)("common.formats.integer",{value:l})),1)]))),128)),t[15]||(t[15]=r()),s("div",lt,[s("dt",null,n(a(i)("data-planes.components.service_traffic_card.2xx")),1),t[11]||(t[11]=r()),s("dd",null,n(a(i)("common.formats.integer",{value:(w=e.traffic.http)==null?void 0:w[`${e.direction}_rq_2xx`]})),1)]),t[16]||(t[16]=r()),(d(!0),c(m,null,q([((E=e.traffic.http)==null?void 0:E[`${e.direction}_rq_3xx`])??0].filter(l=>l!==0),l=>(d(),c("div",{key:l,"data-testid":"rq-3xx"},[s("dt",null,n(a(i)("data-planes.components.service_traffic_card.3xx")),1),t[12]||(t[12]=r()),s("dd",null,n(a(i)("common.formats.integer",{value:l})),1)]))),128)),t[17]||(t[17]=r()),s("div",ct,[s("dt",null,n(a(i)("data-planes.components.service_traffic_card.4xx")),1),t[13]||(t[13]=r()),s("dd",null,n(a(i)("common.formats.integer",{value:(F=e.traffic.http)==null?void 0:F[`${e.direction}_rq_4xx`]})),1)]),t[18]||(t[18]=r()),s("div",ft,[s("dt",null,n(a(i)("data-planes.components.service_traffic_card.5xx")),1),t[14]||(t[14]=r()),s("dd",null,n(a(i)("common.formats.integer",{value:(D=e.traffic.http)==null?void 0:D[`${e.direction}_rq_5xx`]})),1)])],64)):(d(),c(m,{key:3},[s("div",ut,[s("dt",null,n(a(i)("data-planes.components.service_traffic_card.cx")),1),t[19]||(t[19]=r()),s("dd",null,n(a(i)("common.formats.integer",{value:(X=e.traffic.tcp)==null?void 0:X[`${e.direction}_cx_total`]})),1)]),t[22]||(t[22]=r()),typeof((L=e.traffic.tcp)==null?void 0:L[`${e.direction}_cx_tx_bytes_total`])<"u"?(d(),c("div",pt,[s("dt",null,n(a(i)("data-planes.components.service_traffic_card.rx")),1),t[20]||(t[20]=r()),s("dd",null,n(a(k)((P=e.traffic.tcp)==null?void 0:P[`${e.direction}_cx_tx_bytes_total`])),1)])):y("",!0),t[23]||(t[23]=r()),typeof((G=e.traffic.tcp)==null?void 0:G[`${e.direction}_cx_rx_bytes_total`])<"u"?(d(),c("div",_t,[s("dt",null,n(a(i)("data-planes.components.service_traffic_card.tx")),1),t[21]||(t[21]=r()),s("dd",null,n(a(k)((Y=e.traffic.tcp)==null?void 0:Y[`${e.direction}_cx_rx_bytes_total`])),1)])):y("",!0)],64))])):(d(),x(_,{key:2,variant:"line"}))]}),_:3})}}}),kt=T(mt,[["__scopeId","data-v-a3f939b9"]]),vt={class:"body"},xt=M({__name:"ConnectionGroup",props:{type:{}},setup(o){const i=o;return(e,u)=>{const p=C("XCard");return d(),x(p,{class:z(["service-traffic-group",`type-${i.type}`])},{default:g(()=>[s("div",vt,[v(e.$slots,"default",{},void 0,!0)])]),_:3},8,["class"])}}}),Ct=T(xt,[["__scopeId","data-v-25c74403"]]),gt={class:"service-traffic"},yt={class:"actions"},$t=M({__name:"ConnectionTraffic",setup(o){return(i,e)=>(d(),c("div",gt,[s("div",yt,[v(i.$slots,"actions",{},void 0,!0)]),e[0]||(e[0]=r()),K(A,{class:"header"},{title:g(()=>[v(i.$slots,"title",{},void 0,!0)]),_:3}),e[1]||(e[1]=r()),v(i.$slots,"default",{},void 0,!0)]))}}),Tt=T($t,[["__scopeId","data-v-e6bd176c"]]);export{Tt as C,Ct as a,kt as b};
