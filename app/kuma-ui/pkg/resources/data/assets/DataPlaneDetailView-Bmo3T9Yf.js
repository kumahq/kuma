import{d as V,I as Y,c as E,o,a as v,w as i,b as l,t as r,e as B,n as ot,r as Q,p as tt,f as et,g as s,s as st,_ as M,i as x,h as z,k as e,z as lt,j as d,A as c,H as m,J as S,l as ct,F as ut,P as j,a1 as N,a2 as pt,K as X,Y as _t,a3 as at,a4 as ft}from"./index-DKbsM-FP.js";import{f as mt}from"./kong-icons.es321-DO_SOO66.js";import{m as yt}from"./kong-icons.es350-CU-GSgjP.js";import{S as gt}from"./StatusBadge-LDK4At55.js";import{S as vt}from"./SummaryView-GZ3AvkNa.js";import{T as rt}from"./TagList-DfrAkcVd.js";import{_ as ht}from"./SubscriptionList.vue_vue_type_script_setup_true_lang-8rHze9f2.js";import"./AccordionList-BqBffMWx.js";const bt=n=>(tt("data-v-d4a06825"),n=n(),et(),n),kt=["aria-hidden"],xt={key:0,"data-testid":"kui-icon-svg-title"},wt=bt(()=>s("path",{d:"M15 18L21 12L15 6L13.6 7.4L17.2 11H3V13H17.2L13.6 16.6L15 18Z",fill:"currentColor"},null,-1)),$t=V({__name:"ForwardIcon",props:{title:{type:String,required:!1,default:""},color:{type:String,required:!1,default:"currentColor"},display:{type:String,required:!1,default:"block"},decorative:{type:Boolean,required:!1,default:!1},size:{type:[Number,String],required:!1,default:Y,validator:n=>{if(typeof n=="number"&&n>0)return!0;if(typeof n=="string"){const a=String(n).replace(/px/gi,""),t=Number(a);if(t&&!isNaN(t)&&Number.isInteger(t)&&t>0)return!0}return!1}},as:{type:String,required:!1,default:"span"}},setup(n){const a=n,t=E(()=>{if(typeof a.size=="number"&&a.size>0)return`${a.size}px`;if(typeof a.size=="string"){const b=String(a.size).replace(/px/gi,""),y=Number(b);if(y&&!isNaN(y)&&Number.isInteger(y)&&y>0)return`${y}px`}return Y}),h=E(()=>({boxSizing:"border-box",color:a.color,display:a.display,height:t.value,lineHeight:"0",width:t.value}));return(b,y)=>(o(),v(Q(n.as),{"aria-hidden":n.decorative?"true":void 0,class:"kui-icon forward-icon","data-testid":"kui-icon-wrapper-forward-icon",style:ot(h.value)},{default:i(()=>[(o(),l("svg",{"aria-hidden":n.decorative?"true":void 0,"data-testid":"kui-icon-svg-forward-icon",fill:"none",height:"100%",role:"img",viewBox:"0 0 24 24",width:"100%",xmlns:"http://www.w3.org/2000/svg"},[n.title?(o(),l("title",xt,r(n.title),1)):B("",!0),wt],8,kt))]),_:1},8,["aria-hidden","style"]))}}),Ct=st($t,[["__scopeId","data-v-d4a06825"]]),St=n=>(tt("data-v-fccfd59b"),n=n(),et(),n),Bt=["aria-hidden"],It={key:0,"data-testid":"kui-icon-svg-title"},Nt=St(()=>s("path",{d:"M12 21V19H19V5H12V3H19C19.55 3 20.0208 3.19583 20.4125 3.5875C20.8042 3.97917 21 4.45 21 5V19C21 19.55 20.8042 20.0208 20.4125 20.4125C20.0208 20.8042 19.55 21 19 21H12ZM10 17L8.625 15.55L11.175 13H3V11H11.175L8.625 8.45L10 7L15 12L10 17Z",fill:"currentColor"},null,-1)),Tt=V({__name:"GatewayIcon",props:{title:{type:String,required:!1,default:""},color:{type:String,required:!1,default:"currentColor"},display:{type:String,required:!1,default:"block"},decorative:{type:Boolean,required:!1,default:!1},size:{type:[Number,String],required:!1,default:Y,validator:n=>{if(typeof n=="number"&&n>0)return!0;if(typeof n=="string"){const a=String(n).replace(/px/gi,""),t=Number(a);if(t&&!isNaN(t)&&Number.isInteger(t)&&t>0)return!0}return!1}},as:{type:String,required:!1,default:"span"}},setup(n){const a=n,t=E(()=>{if(typeof a.size=="number"&&a.size>0)return`${a.size}px`;if(typeof a.size=="string"){const b=String(a.size).replace(/px/gi,""),y=Number(b);if(y&&!isNaN(y)&&Number.isInteger(y)&&y>0)return`${y}px`}return Y}),h=E(()=>({boxSizing:"border-box",color:a.color,display:a.display,height:t.value,lineHeight:"0",width:t.value}));return(b,y)=>(o(),v(Q(n.as),{"aria-hidden":n.decorative?"true":void 0,class:"kui-icon gateway-icon","data-testid":"kui-icon-wrapper-gateway-icon",style:ot(h.value)},{default:i(()=>[(o(),l("svg",{"aria-hidden":n.decorative?"true":void 0,"data-testid":"kui-icon-svg-gateway-icon",fill:"none",height:"100%",role:"img",viewBox:"0 0 24 24",width:"100%",xmlns:"http://www.w3.org/2000/svg"},[n.title?(o(),l("title",It,r(n.title),1)):B("",!0),Nt],8,Bt))]),_:1},8,["aria-hidden","style"]))}}),qt=st(Tt,[["__scopeId","data-v-fccfd59b"]]),zt=["B","kB","MB","GB","TB","PB","EB","ZB","YB"],Lt=["B","KiB","MiB","GiB","TiB","PiB","EiB","ZiB","YiB"],Dt=["b","kbit","Mbit","Gbit","Tbit","Pbit","Ebit","Zbit","Ybit"],Kt=["b","kibit","Mibit","Gibit","Tibit","Pibit","Eibit","Zibit","Yibit"],nt=(n,a,t)=>{let h=n;return typeof a=="string"||Array.isArray(a)?h=n.toLocaleString(a,t):(a===!0||t!==void 0)&&(h=n.toLocaleString(void 0,t)),h};function Z(n,a){if(!Number.isFinite(n))throw new TypeError(`Expected a finite number, got ${typeof n}: ${n}`);a={bits:!1,binary:!1,space:!0,...a};const t=a.bits?a.binary?Kt:Dt:a.binary?Lt:zt,h=a.space?" ":"";if(a.signed&&n===0)return` 0${h}${t[0]}`;const b=n<0,y=b?"-":a.signed?"+":"";b&&(n=-n);let w;if(a.minimumFractionDigits!==void 0&&(w={minimumFractionDigits:a.minimumFractionDigits}),a.maximumFractionDigits!==void 0&&(w={maximumFractionDigits:a.maximumFractionDigits,...w}),n<1){const D=nt(n,a.locale,w);return y+D+h+t[0]}const C=Math.min(Math.floor(a.binary?Math.log(n)/Math.log(1024):Math.log10(n)/3),t.length-1);n/=(a.binary?1024:1e3)**C,w||(n=n.toPrecision(3));const T=nt(Number(n),a.locale,w),L=t[C];return y+T+h+L}const Vt={},Et={class:"card"},Mt={class:"title"},Ht={class:"body"};function Rt(n,a){const t=x("KCard");return o(),v(t,{class:"data-card"},{default:i(()=>[s("dl",null,[s("div",Et,[s("dt",Mt,[z(n.$slots,"title",{},void 0,!0)]),e(),s("dd",Ht,[z(n.$slots,"default",{},void 0,!0)])])])]),_:3})}const dt=M(Vt,[["render",Rt],["__scopeId","data-v-3f9a3cf3"]]),Pt={class:"title"},At={key:0},Ut={"data-testid":"grpc-success"},Ft={"data-testid":"grpc-failure"},Ot={"data-testid":"rq-2xx"},Gt={"data-testid":"rq-4xx"},Zt={"data-testid":"rq-5xx"},Yt={"data-testid":"connections-total"},jt={key:0,"data-testid":"bytes-received"},Xt={key:1,"data-testid":"bytes-sent"},Jt=V({__name:"ConnectionCard",props:{protocol:{},service:{default:""},traffic:{default:void 0},direction:{default:"downstream"}},setup(n){const{t:a}=lt(),t=n,h=b=>{const y=b.target;if(b.isTrusted&&y.nodeName.toLowerCase()!=="a"){const w=y.closest(".service-traffic-card, a");if(w){const C=w.nodeName.toLowerCase()==="a"?w:w.querySelector("[data-action]");C!==null&&"click"in C&&typeof C.click=="function"&&C.click()}}};return(b,y)=>{const w=x("KBadge"),C=x("KSkeletonBox");return o(),v(dt,{class:"service-traffic-card",onClick:h},{title:i(()=>[t.service.length>0?(o(),v(rt,{key:0,tags:[{label:"kuma.io/service",value:t.service}]},null,8,["tags"])):B("",!0),e(),s("div",Pt,[d(w,{class:"protocol",appearance:t.protocol==="passthrough"?"success":"info"},{default:i(()=>[e(r(c(a)(`data-planes.components.service_traffic_card.protocol.${t.protocol}`,{},{defaultMessage:c(a)(`http.api.value.${t.protocol}`)})),1)]),_:1},8,["appearance"]),e(),z(b.$slots,"default",{},void 0,!0)])]),default:i(()=>{var T,L,D,H,R,P,A,U,F,O,$,_;return[e(),t.traffic?(o(),l("dl",At,[t.protocol==="passthrough"?(o(!0),l(m,{key:0},S([["http","tcp"].reduce((p,q)=>{var u;const G=t.direction;return Object.entries(((u=t.traffic)==null?void 0:u[q])||{}).reduce((f,[g,k])=>[`${G}_cx_tx_bytes_total`,`${G}_cx_rx_bytes_total`].includes(g)?{...f,[g]:k+(f[g]??0)}:f,p)},{})],(p,q)=>(o(),l(m,{key:q},[s("div",null,[s("dt",null,r(c(a)("data-planes.components.service_traffic_card.tx")),1),e(),s("dd",null,r(c(Z)(p.downstream_cx_rx_bytes_total??0)),1)]),e(),s("div",null,[s("dt",null,r(c(a)("data-planes.components.service_traffic_card.rx")),1),e(),s("dd",null,r(c(Z)(p.downstream_cx_tx_bytes_total??0)),1)])],64))),128)):t.protocol==="grpc"?(o(),l(m,{key:1},[s("div",Ut,[s("dt",null,r(c(a)("data-planes.components.service_traffic_card.grpc_success")),1),e(),s("dd",null,r(c(a)("common.formats.integer",{value:(T=t.traffic.grpc)==null?void 0:T.success})),1)]),e(),s("div",Ft,[s("dt",null,r(c(a)("data-planes.components.service_traffic_card.grpc_failure")),1),e(),s("dd",null,r(c(a)("common.formats.integer",{value:(L=t.traffic.grpc)==null?void 0:L.failure})),1)])],64)):t.protocol.startsWith("http")?(o(),l(m,{key:2},[(o(!0),l(m,null,S([((D=t.traffic.http)==null?void 0:D[`${t.direction}_rq_1xx`])??0].filter(p=>p!==0),p=>(o(),l("div",{key:p,"data-testid":"rq-1xx"},[s("dt",null,r(c(a)("data-planes.components.service_traffic_card.1xx")),1),e(),s("dd",null,r(c(a)("common.formats.integer",{value:p})),1)]))),128)),e(),s("div",Ot,[s("dt",null,r(c(a)("data-planes.components.service_traffic_card.2xx")),1),e(),s("dd",null,r(c(a)("common.formats.integer",{value:(H=t.traffic.http)==null?void 0:H[`${t.direction}_rq_2xx`]})),1)]),e(),(o(!0),l(m,null,S([((R=t.traffic.http)==null?void 0:R[`${t.direction}_rq_3xx`])??0].filter(p=>p!==0),p=>(o(),l("div",{key:p,"data-testid":"rq-3xx"},[s("dt",null,r(c(a)("data-planes.components.service_traffic_card.3xx")),1),e(),s("dd",null,r(c(a)("common.formats.integer",{value:p})),1)]))),128)),e(),s("div",Gt,[s("dt",null,r(c(a)("data-planes.components.service_traffic_card.4xx")),1),e(),s("dd",null,r(c(a)("common.formats.integer",{value:(P=t.traffic.http)==null?void 0:P[`${t.direction}_rq_4xx`]})),1)]),e(),s("div",Zt,[s("dt",null,r(c(a)("data-planes.components.service_traffic_card.5xx")),1),e(),s("dd",null,r(c(a)("common.formats.integer",{value:(A=t.traffic.http)==null?void 0:A[`${t.direction}_rq_5xx`]})),1)])],64)):(o(),l(m,{key:3},[s("div",Yt,[s("dt",null,r(c(a)("data-planes.components.service_traffic_card.cx")),1),e(),s("dd",null,r(c(a)("common.formats.integer",{value:(U=t.traffic.tcp)==null?void 0:U[`${t.direction}_cx_total`]})),1)]),e(),typeof((F=t.traffic.tcp)==null?void 0:F[`${t.direction}_cx_tx_bytes_total`])<"u"?(o(),l("div",jt,[s("dt",null,r(c(a)("data-planes.components.service_traffic_card.rx")),1),e(),s("dd",null,r(c(Z)((O=t.traffic.tcp)==null?void 0:O[`${t.direction}_cx_tx_bytes_total`])),1)])):B("",!0),e(),typeof(($=t.traffic.tcp)==null?void 0:$[`${t.direction}_cx_rx_bytes_total`])<"u"?(o(),l("div",Xt,[s("dt",null,r(c(a)("data-planes.components.service_traffic_card.tx")),1),e(),s("dd",null,r(c(Z)((_=t.traffic.tcp)==null?void 0:_[`${t.direction}_cx_rx_bytes_total`])),1)])):B("",!0)],64))])):(o(),v(C,{key:1,width:"10"}))]}),_:3})}}}),J=M(Jt,[["__scopeId","data-v-f7ef7711"]]),Wt={class:"body"},Qt=V({__name:"ConnectionGroup",props:{type:{}},setup(n){const a=n;return(t,h)=>{const b=x("KCard");return o(),v(b,{class:ct(["service-traffic-group",`type-${a.type}`])},{default:i(()=>[s("div",Wt,[z(t.$slots,"default",{},void 0,!0)])]),_:3},8,["class"])}}}),W=M(Qt,[["__scopeId","data-v-9402d5d1"]]),te={class:"service-traffic"},ee={class:"actions"},ae=V({__name:"ConnectionTraffic",setup(n){return(a,t)=>(o(),l("div",te,[s("div",ee,[z(a.$slots,"actions",{},void 0,!0)]),e(),d(dt,{class:"header"},{title:i(()=>[z(a.$slots,"title",{},void 0,!0)]),_:3}),e(),z(a.$slots,"default",{},void 0,!0)]))}}),it=M(ae,[["__scopeId","data-v-e6bd176c"]]),ne=n=>(tt("data-v-0ac1ecac"),n=n(),et(),n),ie={"data-testid":"dataplane-warnings"},oe=["data-testid","innerHTML"],se={key:0,"data-testid":"warning-stats-loading"},re={class:"stack","data-testid":"dataplane-details"},de={class:"columns"},le={class:"status-with-reason"},ce={class:"columns"},ue=ne(()=>s("span",null,"Outbounds",-1)),pe={"data-testid":"dataplane-mtls"},_e={class:"columns"},fe=["innerHTML"],me={key:0,"data-testid":"dataplane-subscriptions"},ye=V({__name:"DataPlaneDetailView",props:{data:{}},setup(n){const a=ut(),t=n,h=E(()=>t.data.warnings.concat(...t.data.isCertExpired?[{kind:"CERT_EXPIRED"}]:[]));return(b,y)=>{const w=x("KTooltip"),C=x("DataCollection"),T=x("KCard"),L=x("XAction"),D=x("KInputSwitch"),H=x("KButton"),R=x("RouterLink"),P=x("RouterView"),A=x("KAlert"),U=x("AppView"),F=x("DataSource"),O=x("RouteView");return o(),v(O,{params:{mesh:"",dataPlane:"",inactive:!1},name:"data-plane-detail-view"},{default:i(({route:$,t:_})=>[d(F,{src:`/meshes/${$.params.mesh}/dataplanes/${$.params.dataPlane}/stats/${t.data.dataplane.networking.inboundAddress}`},{default:i(({data:p,error:q,refresh:G})=>[d(U,null,j({default:i(()=>[e(),s("div",re,[d(T,null,{default:i(()=>[s("div",de,[d(N,null,{title:i(()=>[e(r(_("http.api.property.status")),1)]),body:i(()=>[s("div",le,[d(gt,{status:t.data.status},null,8,["status"]),e(),t.data.dataplaneType==="standard"?(o(),v(C,{key:0,items:t.data.dataplane.networking.inbounds,predicate:u=>!u.health.ready,empty:!1},{default:i(({items:u})=>[d(w,{class:"reason-tooltip"},{content:i(()=>[s("ul",null,[(o(!0),l(m,null,S(u,f=>(o(),l("li",{key:`${f.service}:${f.port}`},r(_("data-planes.routes.item.unhealthy_inbound",{service:f.service,port:f.port})),1))),128))])]),default:i(()=>[d(c(mt),{color:c(pt),size:c(X)},null,8,["color","size"]),e()]),_:2},1024)]),_:2},1032,["items","predicate"])):B("",!0)])]),_:2},1024),e(),d(N,null,{title:i(()=>[e(`
                  Type
                `)]),body:i(()=>[e(r(_(`data-planes.type.${t.data.dataplaneType}`)),1)]),_:2},1024),e(),t.data.namespace.length>0?(o(),v(N,{key:0},{title:i(()=>[e(`
                  Namespace
                `)]),body:i(()=>[e(r(t.data.namespace),1)]),_:1})):B("",!0),e(),d(N,null,{title:i(()=>[e(r(_("data-planes.routes.item.last_updated")),1)]),body:i(()=>[e(r(_("common.formats.datetime",{value:Date.parse(t.data.modificationTime)})),1)]),_:2},1024),e(),t.data.dataplane.networking.gateway?(o(),l(m,{key:1},[d(N,null,{title:i(()=>[e(r(_("http.api.property.tags")),1)]),body:i(()=>[d(rt,{tags:t.data.dataplane.networking.gateway.tags},null,8,["tags"])]),_:2},1024),e(),d(N,null,{title:i(()=>[e(r(_("http.api.property.address")),1)]),body:i(()=>[d(_t,{text:`${t.data.dataplane.networking.address}`},null,8,["text"])]),_:2},1024)],64)):B("",!0)])]),_:2},1024),e(),d(T,{class:"traffic","data-testid":"dataplane-traffic"},{default:i(()=>[s("div",ce,[d(it,null,{title:i(()=>[d(c(Ct),{display:"inline-block",decorative:"",size:c(X)},null,8,["size"]),e(`
                  Inbounds
                `)]),default:i(()=>[e(),d(W,{type:"inbound","data-testid":"dataplane-inbounds"},{default:i(()=>[(o(!0),l(m,null,S([t.data.dataplane.networking.type==="gateway"?Object.entries((p==null?void 0:p.inbounds)??{}).reduce((u,[f,g])=>{var I;const k=f.split("_").at(-1);return k===(((I=t.data.dataplane.networking.admin)==null?void 0:I.port)??"9901")?u:u.concat([{...t.data.dataplane.networking.inbounds[0],name:f,port:Number(k),protocol:["http","tcp"].find(K=>typeof g[K]<"u")??"tcp",addressPort:`${t.data.dataplane.networking.inbounds[0].address}:${k}`}])},[]):t.data.dataplane.networking.inbounds],u=>(o(),v(C,{key:u,items:u,predicate:f=>f.port!==49151},j({default:i(({items:f})=>[(o(!0),l(m,null,S(f,g=>(o(),l(m,{key:`${g.name}`},[(o(!0),l(m,null,S([p==null?void 0:p.inbounds[g.name]],k=>(o(),v(J,{key:k,"data-testid":"dataplane-inbound",protocol:g.protocol,service:g.tags["kuma.io/service"],traffic:typeof q>"u"?k:{name:"",protocol:g.protocol,port:`${g.port}`}},{default:i(()=>[d(L,{"data-action":"",to:{name:(I=>I.includes("bound")?I.replace("-outbound-","-inbound-"):"connection-inbound-summary-overview-view")(String(c(a).name)),params:{connection:g.name},query:{inactive:$.params.inactive}}},{default:i(()=>[e(r(g.name.replace("localhost","").replace("_",":")),1)]),_:2},1032,["to"])]),_:2},1032,["protocol","service","traffic"]))),128))],64))),128))]),_:2},[t.data.dataplaneType==="delegated"?{name:"empty",fn:i(()=>[d(at,null,{default:i(()=>[e(`
                          This proxy is a delegated gateway therefore `+r(_("common.product.name"))+` does not have any visibility into inbounds for this gateway
                        `,1)]),_:2},1024)]),key:"0"}:void 0]),1032,["items","predicate"]))),128))]),_:2},1024)]),_:2},1024),e(),d(it,null,j({title:i(()=>[d(c(qt),{display:"inline-block",decorative:"",size:c(X)},null,8,["size"]),e(),ue]),default:i(()=>[e(),e(),typeof q>"u"?(o(),l(m,{key:0},[typeof p>"u"?(o(),v(ft,{key:0})):(o(),l(m,{key:1},S(["upstream"],u=>(o(),l(m,{key:u},[d(W,{type:"passthrough"},{default:i(()=>[d(J,{protocol:"passthrough",traffic:p.passthrough},{default:i(()=>[e(`
                          Non mesh traffic
                        `)]),_:2},1032,["traffic"])]),_:2},1024),e(),d(C,{predicate:$.params.inactive?void 0:([f,g])=>{var k,I;return((typeof g.tcp<"u"?(k=g.tcp)==null?void 0:k[`${u}_cx_rx_bytes_total`]:(I=g.http)==null?void 0:I[`${u}_rq_total`])??0)>0},items:Object.entries(p.outbounds)},{default:i(({items:f})=>[f.length>0?(o(),v(W,{key:0,type:"outbound","data-testid":"dataplane-outbounds"},{default:i(()=>[(o(),l(m,null,S([/-([a-f0-9]){16}$/],g=>(o(),l(m,{key:g},[(o(!0),l(m,null,S(f,([k,I])=>(o(),v(J,{key:`${k}`,"data-testid":"dataplane-outbound",protocol:["grpc","http","tcp"].find(K=>typeof I[K]<"u")??"tcp",traffic:I,service:k.replace(g,""),direction:u},{default:i(()=>[d(R,{"data-action":"",to:{name:(K=>K.includes("bound")?K.replace("-inbound-","-outbound-"):"connection-outbound-summary-overview-view")(String(c(a).name)),params:{connection:k},query:{inactive:$.params.inactive?null:void 0}}},{default:i(()=>[e(r(k),1)]),_:2},1032,["to"])]),_:2},1032,["protocol","traffic","service","direction"]))),128))],64))),64))]),_:2},1024)):B("",!0)]),_:2},1032,["predicate","items"])],64))),64))],64)):(o(),v(at,{key:1}))]),_:2},[p?{name:"actions",fn:i(()=>[d(D,{modelValue:$.params.inactive,"onUpdate:modelValue":u=>$.params.inactive=u,"data-testid":"dataplane-outbounds-inactive-toggle"},{label:i(()=>[e(`
                      Show inactive
                    `)]),_:2},1032,["modelValue","onUpdate:modelValue"]),e(),d(H,{appearance:"primary",onClick:G},{default:i(()=>[d(c(yt)),e(`

                    Refresh
                  `)]),_:2},1032,["onClick"])]),key:"0"}:void 0]),1024)])]),_:2},1024),e(),d(P,null,{default:i(u=>[u.route.name!==$.name?(o(),v(vt,{key:0,width:"670px",onClose:function(){$.replace({name:"data-plane-detail-view",params:{mesh:$.params.mesh,dataPlane:$.params.dataPlane},query:{inactive:$.params.inactive?null:void 0}})}},{default:i(()=>[(o(),v(Q(u.Component),{data:u.route.name.includes("-inbound-")?t.data.dataplane.networking.inbounds:(p==null?void 0:p.outbounds)||{},"dataplane-overview":t.data},null,8,["data","dataplane-overview"]))]),_:2},1032,["onClose"])):B("",!0)]),_:2},1024),e(),s("div",pe,[s("h2",null,r(_("data-planes.routes.item.mtls.title")),1),e(),t.data.dataplaneInsight.mTLS?(o(!0),l(m,{key:0},S([t.data.dataplaneInsight.mTLS],u=>(o(),v(T,{key:u,class:"mt-4"},{default:i(()=>[s("div",_e,[d(N,null,{title:i(()=>[e(r(_("data-planes.routes.item.mtls.expiration_time.title")),1)]),body:i(()=>[e(r(_("common.formats.datetime",{value:Date.parse(u.certificateExpirationTime)})),1)]),_:2},1024),e(),d(N,null,{title:i(()=>[e(r(_("data-planes.routes.item.mtls.generation_time.title")),1)]),body:i(()=>[e(r(_("common.formats.datetime",{value:Date.parse(u.lastCertificateRegeneration)})),1)]),_:2},1024),e(),d(N,null,{title:i(()=>[e(r(_("data-planes.routes.item.mtls.regenerations.title")),1)]),body:i(()=>[e(r(_("common.formats.integer",{value:u.certificateRegenerations})),1)]),_:2},1024),e(),d(N,null,{title:i(()=>[e(r(_("data-planes.routes.item.mtls.issued_backend.title")),1)]),body:i(()=>[e(r(u.issuedBackend),1)]),_:2},1024),e(),d(N,null,{title:i(()=>[e(r(_("data-planes.routes.item.mtls.supported_backends.title")),1)]),body:i(()=>[s("ul",null,[(o(!0),l(m,null,S(u.supportedBackends,f=>(o(),l("li",{key:f},r(f),1))),128))])]),_:2},1024)])]),_:2},1024))),128)):(o(),v(A,{key:1,class:"mt-4",appearance:"warning"},{default:i(()=>[s("div",{innerHTML:_("data-planes.routes.item.mtls.disabled")},null,8,fe)]),_:2},1024))]),e(),t.data.dataplaneInsight.subscriptions.length>0?(o(),l("div",me,[s("h2",null,r(_("data-planes.routes.item.subscriptions.title")),1),e(),d(T,{class:"mt-4"},{default:i(()=>[d(ht,{subscriptions:t.data.dataplaneInsight.subscriptions},null,8,["subscriptions"])]),_:1})])):B("",!0)])]),_:2},[h.value.length>0||q?{name:"notifications",fn:i(()=>[s("ul",ie,[(o(!0),l(m,null,S(h.value,u=>(o(),l("li",{key:u.kind,"data-testid":`warning-${u.kind}`,innerHTML:_(`common.warnings.${u.kind}`,u.payload)},null,8,oe))),128)),e(),q?(o(),l("li",se,[e(`
              The below view is not enhanced with runtime stats (Error loading stats: `),s("strong",null,r(q.toString()),1),e(`)
            `)])):B("",!0)])]),key:"0"}:void 0]),1024)]),_:2},1032,["src"])]),_:1})}}}),Ce=M(ye,[["__scopeId","data-v-0ac1ecac"]]);export{Ce as default};
