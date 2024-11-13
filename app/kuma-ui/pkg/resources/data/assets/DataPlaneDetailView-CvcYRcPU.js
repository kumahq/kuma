import{d as R,I as Y,D as P,o as i,k as h,w as n,c,E as st,C as tt,m as A,e as C,i as s,r as M,b as e,g as dt,l as D,a as l,t as o,j as p,F as y,G as N,n as lt,B as ct,Q as j,P as z,S as ut,a0 as pt,K as J,$ as _t,a1 as ft,A as mt,H as gt,J as yt}from"./index-Dxk6JoIL.js";import{q as vt}from"./kong-icons.es678-BbdCwclO.js";import{S as ht}from"./SummaryView-1NztUQvH.js";import{T as ot}from"./TagList-CSKNWVa1.js";const bt=["aria-hidden"],et='<path d="M15 18L21 12L15 6L13.6 7.4L17.2 11H3V13H17.2L13.6 16.6L15 18Z" fill="currentColor"/>',kt=R({__name:"ForwardIcon",props:{title:{type:String,required:!1,default:""},color:{type:String,required:!1,default:"currentColor"},display:{type:String,required:!1,default:"block"},decorative:{type:Boolean,required:!1,default:!1},size:{type:[Number,String],required:!1,default:Y,validator:r=>{if(typeof r=="number"&&r>0)return!0;if(typeof r=="string"){const a=String(r).replace(/px/gi,""),t=Number(a);if(t&&!isNaN(t)&&Number.isInteger(t)&&t>0)return!0}return!1}},as:{type:String,required:!1,default:"span"},staticIds:{type:Boolean,default:!1}},setup(r){const a=r,t=P(()=>{if(typeof a.size=="number"&&a.size>0)return`${a.size}px`;if(typeof a.size=="string"){const g=String(a.size).replace(/px/gi,""),u=Number(g);if(u&&!isNaN(u)&&Number.isInteger(u)&&u>0)return`${u}px`}return Y}),k=P(()=>({boxSizing:"border-box",color:a.color,display:a.display,flexShrink:"0",height:t.value,lineHeight:"0",width:t.value,pointerEvents:a.decorative?"none":void 0})),I=g=>{const u={},w=Math.random().toString(36).substring(2,12);return g.replace(/id="([^"]+)"/g,(T,S)=>{const V=`${w}-${S}`;return u[S]=V,`id="${V}"`}).replace(/#([^\s^")]+)/g,(T,S)=>u[S]?`#${u[S]}`:T)},L=g=>{const u={"<":"&lt;",">":"&gt;",'"':"&quot;","'":"&#039;","`":"&#039;"};return g.replace(/[<>"'`]/g,w=>u[w])},b=`${a.title?`<title data-testid="kui-icon-svg-title">${L(a.title)}</title>`:""}${a.staticIds?et:I(et)}`;return(g,u)=>(i(),h(tt(r.as),{"aria-hidden":r.decorative?"true":void 0,class:"kui-icon forward-icon","data-testid":"kui-icon-wrapper-forward-icon",style:st(k.value),tabindex:r.decorative?"-1":void 0},{default:n(()=>[(i(),c("svg",{"aria-hidden":r.decorative?"true":void 0,"data-testid":"kui-icon-svg-forward-icon",fill:"none",height:"100%",role:"img",viewBox:"0 0 24 24",width:"100%",xmlns:"http://www.w3.org/2000/svg",innerHTML:b},null,8,bt))]),_:1},8,["aria-hidden","style","tabindex"]))}}),xt=["aria-hidden"],at='<path d="M12 21V19H19V5H12V3H19C19.55 3 20.0208 3.19583 20.4125 3.5875C20.8042 3.97917 21 4.45 21 5V19C21 19.55 20.8042 20.0208 20.4125 20.4125C20.0208 20.8042 19.55 21 19 21H12ZM10 17L8.625 15.55L11.175 13H3V11H11.175L8.625 8.45L10 7L15 12L10 17Z" fill="currentColor"/>',$t=R({__name:"GatewayIcon",props:{title:{type:String,required:!1,default:""},color:{type:String,required:!1,default:"currentColor"},display:{type:String,required:!1,default:"block"},decorative:{type:Boolean,required:!1,default:!1},size:{type:[Number,String],required:!1,default:Y,validator:r=>{if(typeof r=="number"&&r>0)return!0;if(typeof r=="string"){const a=String(r).replace(/px/gi,""),t=Number(a);if(t&&!isNaN(t)&&Number.isInteger(t)&&t>0)return!0}return!1}},as:{type:String,required:!1,default:"span"},staticIds:{type:Boolean,default:!1}},setup(r){const a=r,t=P(()=>{if(typeof a.size=="number"&&a.size>0)return`${a.size}px`;if(typeof a.size=="string"){const g=String(a.size).replace(/px/gi,""),u=Number(g);if(u&&!isNaN(u)&&Number.isInteger(u)&&u>0)return`${u}px`}return Y}),k=P(()=>({boxSizing:"border-box",color:a.color,display:a.display,flexShrink:"0",height:t.value,lineHeight:"0",width:t.value,pointerEvents:a.decorative?"none":void 0})),I=g=>{const u={},w=Math.random().toString(36).substring(2,12);return g.replace(/id="([^"]+)"/g,(T,S)=>{const V=`${w}-${S}`;return u[S]=V,`id="${V}"`}).replace(/#([^\s^")]+)/g,(T,S)=>u[S]?`#${u[S]}`:T)},L=g=>{const u={"<":"&lt;",">":"&gt;",'"':"&quot;","'":"&#039;","`":"&#039;"};return g.replace(/[<>"'`]/g,w=>u[w])},b=`${a.title?`<title data-testid="kui-icon-svg-title">${L(a.title)}</title>`:""}${a.staticIds?at:I(at)}`;return(g,u)=>(i(),h(tt(r.as),{"aria-hidden":r.decorative?"true":void 0,class:"kui-icon gateway-icon","data-testid":"kui-icon-wrapper-gateway-icon",style:st(k.value),tabindex:r.decorative?"-1":void 0},{default:n(()=>[(i(),c("svg",{"aria-hidden":r.decorative?"true":void 0,"data-testid":"kui-icon-svg-gateway-icon",fill:"none",height:"100%",role:"img",viewBox:"0 0 24 24",width:"100%",xmlns:"http://www.w3.org/2000/svg",innerHTML:b},null,8,xt))]),_:1},8,["aria-hidden","style","tabindex"]))}}),wt=["B","kB","MB","GB","TB","PB","EB","ZB","YB"],St=["B","KiB","MiB","GiB","TiB","PiB","EiB","ZiB","YiB"],Ct=["b","kbit","Mbit","Gbit","Tbit","Pbit","Ebit","Zbit","Ybit"],It=["b","kibit","Mibit","Gibit","Tibit","Pibit","Eibit","Zibit","Yibit"],nt=(r,a,t)=>{let k=r;return typeof a=="string"||Array.isArray(a)?k=r.toLocaleString(a,t):(a===!0||t!==void 0)&&(k=r.toLocaleString(void 0,t)),k};function Z(r,a){if(!Number.isFinite(r))throw new TypeError(`Expected a finite number, got ${typeof r}: ${r}`);a={bits:!1,binary:!1,space:!0,...a};const t=a.bits?a.binary?It:Ct:a.binary?St:wt,k=a.space?" ":"";if(a.signed&&r===0)return` 0${k}${t[0]}`;const I=r<0,L=I?"-":a.signed?"+":"";I&&(r=-r);let b;if(a.minimumFractionDigits!==void 0&&(b={minimumFractionDigits:a.minimumFractionDigits}),a.maximumFractionDigits!==void 0&&(b={maximumFractionDigits:a.maximumFractionDigits,...b}),r<1){const T=nt(r,a.locale,b);return L+T+k+t[0]}const g=Math.min(Math.floor(a.binary?Math.log(r)/Math.log(1024):Math.log10(r)/3),t.length-1);r/=(a.binary?1024:1e3)**g,b||(r=r.toPrecision(3));const u=nt(Number(r),a.locale,b),w=t[g];return L+u+k+w}const Bt={},Nt={class:"card"},Tt={class:"title"},qt={class:"body"};function zt(r,a){const t=C("KCard");return i(),h(t,{class:"data-card"},{default:n(()=>[s("dl",null,[s("div",Nt,[s("dt",Tt,[M(r.$slots,"title",{},void 0,!0)]),e(),s("dd",qt,[M(r.$slots,"default",{},void 0,!0)])])])]),_:3})}const rt=A(Bt,[["render",zt],["__scopeId","data-v-3f9a3cf3"]]),Dt={class:"title"},Lt={key:0},Vt={"data-testid":"grpc-success"},Et={"data-testid":"grpc-failure"},Mt={"data-testid":"rq-2xx"},Ht={"data-testid":"rq-4xx"},Rt={"data-testid":"rq-5xx"},Kt={"data-testid":"connections-total"},Pt={key:0,"data-testid":"bytes-received"},At={key:1,"data-testid":"bytes-sent"},Ut=R({__name:"ConnectionCard",props:{protocol:{},service:{default:""},traffic:{default:void 0},direction:{default:"downstream"}},setup(r){const{t:a}=dt(),t=r,k=I=>{const L=I.target;if(I.isTrusted&&L.nodeName.toLowerCase()!=="a"){const b=L.closest(".service-traffic-card, a");if(b){const g=b.nodeName.toLowerCase()==="a"?b:b.querySelector("[data-action]");g!==null&&"click"in g&&typeof g.click=="function"&&g.click()}}};return(I,L)=>{const b=C("XBadge"),g=C("KSkeletonBox");return i(),h(rt,{class:"service-traffic-card",onClick:k},{title:n(()=>[t.service.length>0?(i(),h(ot,{key:0,tags:[{label:"kuma.io/service",value:t.service}]},null,8,["tags"])):D("",!0),e(),s("div",Dt,[l(b,{class:"protocol",appearance:t.protocol==="passthrough"?"success":"info"},{default:n(()=>[e(o(p(a)(`data-planes.components.service_traffic_card.protocol.${t.protocol}`,{},{defaultMessage:p(a)(`http.api.value.${t.protocol}`)})),1)]),_:1},8,["appearance"]),e(),M(I.$slots,"default",{},void 0,!0)])]),default:n(()=>{var u,w,T,S,V,U,X,F,G,x,_,K;return[e(),t.traffic?(i(),c("dl",Lt,[t.protocol==="passthrough"?(i(!0),c(y,{key:0},N([["http","tcp"].reduce((v,$)=>{var O;const E=t.direction;return Object.entries(((O=t.traffic)==null?void 0:O[$])||{}).reduce((d,[f,m])=>[`${E}_cx_tx_bytes_total`,`${E}_cx_rx_bytes_total`].includes(f)?{...d,[f]:m+(d[f]??0)}:d,v)},{})],(v,$)=>(i(),c(y,{key:$},[s("div",null,[s("dt",null,o(p(a)("data-planes.components.service_traffic_card.tx")),1),e(),s("dd",null,o(p(Z)(v.downstream_cx_rx_bytes_total??0)),1)]),e(),s("div",null,[s("dt",null,o(p(a)("data-planes.components.service_traffic_card.rx")),1),e(),s("dd",null,o(p(Z)(v.downstream_cx_tx_bytes_total??0)),1)])],64))),128)):t.protocol==="grpc"?(i(),c(y,{key:1},[s("div",Vt,[s("dt",null,o(p(a)("data-planes.components.service_traffic_card.grpc_success")),1),e(),s("dd",null,o(p(a)("common.formats.integer",{value:(u=t.traffic.grpc)==null?void 0:u.success})),1)]),e(),s("div",Et,[s("dt",null,o(p(a)("data-planes.components.service_traffic_card.grpc_failure")),1),e(),s("dd",null,o(p(a)("common.formats.integer",{value:(w=t.traffic.grpc)==null?void 0:w.failure})),1)])],64)):t.protocol.startsWith("http")?(i(),c(y,{key:2},[(i(!0),c(y,null,N([((T=t.traffic.http)==null?void 0:T[`${t.direction}_rq_1xx`])??0].filter(v=>v!==0),v=>(i(),c("div",{key:v,"data-testid":"rq-1xx"},[s("dt",null,o(p(a)("data-planes.components.service_traffic_card.1xx")),1),e(),s("dd",null,o(p(a)("common.formats.integer",{value:v})),1)]))),128)),e(),s("div",Mt,[s("dt",null,o(p(a)("data-planes.components.service_traffic_card.2xx")),1),e(),s("dd",null,o(p(a)("common.formats.integer",{value:(S=t.traffic.http)==null?void 0:S[`${t.direction}_rq_2xx`]})),1)]),e(),(i(!0),c(y,null,N([((V=t.traffic.http)==null?void 0:V[`${t.direction}_rq_3xx`])??0].filter(v=>v!==0),v=>(i(),c("div",{key:v,"data-testid":"rq-3xx"},[s("dt",null,o(p(a)("data-planes.components.service_traffic_card.3xx")),1),e(),s("dd",null,o(p(a)("common.formats.integer",{value:v})),1)]))),128)),e(),s("div",Ht,[s("dt",null,o(p(a)("data-planes.components.service_traffic_card.4xx")),1),e(),s("dd",null,o(p(a)("common.formats.integer",{value:(U=t.traffic.http)==null?void 0:U[`${t.direction}_rq_4xx`]})),1)]),e(),s("div",Rt,[s("dt",null,o(p(a)("data-planes.components.service_traffic_card.5xx")),1),e(),s("dd",null,o(p(a)("common.formats.integer",{value:(X=t.traffic.http)==null?void 0:X[`${t.direction}_rq_5xx`]})),1)])],64)):(i(),c(y,{key:3},[s("div",Kt,[s("dt",null,o(p(a)("data-planes.components.service_traffic_card.cx")),1),e(),s("dd",null,o(p(a)("common.formats.integer",{value:(F=t.traffic.tcp)==null?void 0:F[`${t.direction}_cx_total`]})),1)]),e(),typeof((G=t.traffic.tcp)==null?void 0:G[`${t.direction}_cx_tx_bytes_total`])<"u"?(i(),c("div",Pt,[s("dt",null,o(p(a)("data-planes.components.service_traffic_card.rx")),1),e(),s("dd",null,o(p(Z)((x=t.traffic.tcp)==null?void 0:x[`${t.direction}_cx_tx_bytes_total`])),1)])):D("",!0),e(),typeof((_=t.traffic.tcp)==null?void 0:_[`${t.direction}_cx_rx_bytes_total`])<"u"?(i(),c("div",At,[s("dt",null,o(p(a)("data-planes.components.service_traffic_card.tx")),1),e(),s("dd",null,o(p(Z)((K=t.traffic.tcp)==null?void 0:K[`${t.direction}_cx_rx_bytes_total`])),1)])):D("",!0)],64))])):(i(),h(g,{key:1,width:"10"}))]}),_:3})}}}),Q=A(Ut,[["__scopeId","data-v-82875ef3"]]),Xt={class:"body"},Ft=R({__name:"ConnectionGroup",props:{type:{}},setup(r){const a=r;return(t,k)=>{const I=C("KCard");return i(),h(I,{class:lt(["service-traffic-group",`type-${a.type}`])},{default:n(()=>[s("div",Xt,[M(t.$slots,"default",{},void 0,!0)])]),_:3},8,["class"])}}}),W=A(Ft,[["__scopeId","data-v-9402d5d1"]]),Gt={class:"service-traffic"},Ot={class:"actions"},Zt=R({__name:"ConnectionTraffic",setup(r){return(a,t)=>(i(),c("div",Gt,[s("div",Ot,[M(a.$slots,"actions",{},void 0,!0)]),e(),l(rt,{class:"header"},{title:n(()=>[M(a.$slots,"title",{},void 0,!0)]),_:3}),e(),M(a.$slots,"default",{},void 0,!0)]))}}),it=A(Zt,[["__scopeId","data-v-e6bd176c"]]),Yt=r=>(gt("data-v-5bed4307"),r=r(),yt(),r),jt={"data-testid":"dataplane-warnings"},Jt=["data-testid","innerHTML"],Qt={key:0,"data-testid":"warning-stats-loading"},Wt={class:"stack","data-testid":"dataplane-details"},te={class:"stack"},ee={class:"columns"},ae={class:"status-with-reason"},ne={class:"columns"},ie={class:"columns"},se=Yt(()=>s("span",null,"Outbounds",-1)),oe={"data-testid":"dataplane-mtls"},re={class:"columns"},de=["innerHTML"],le={key:0,"data-testid":"dataplane-subscriptions"},ce=R({__name:"DataPlaneDetailView",props:{data:{},mesh:{}},setup(r){const a=ct(),t=r,k=P(()=>t.data.warnings.concat(...t.data.isCertExpired?[{kind:"CERT_EXPIRED"}]:[]));return(I,L)=>{const b=C("KTooltip"),g=C("DataCollection"),u=C("XAction"),w=C("KCard"),T=C("XEmptyState"),S=C("XInputSwitch"),V=C("RouterView"),U=C("XAlert"),X=C("AppView"),F=C("DataSource"),G=C("RouteView");return i(),h(G,{params:{mesh:"",dataPlane:"",subscription:"",inactive:!1},name:"data-plane-detail-view"},{default:n(({route:x,t:_,can:K,me:v})=>[l(F,{src:`/meshes/${x.params.mesh}/dataplanes/${x.params.dataPlane}/stats/${t.data.dataplane.networking.inboundAddress}`},{default:n(({data:$,error:E,refresh:O})=>[l(X,null,j({default:n(()=>[e(),s("div",Wt,[l(w,null,{default:n(()=>[s("div",te,[s("div",ee,[l(z,null,{title:n(()=>[e(o(_("http.api.property.status")),1)]),body:n(()=>[s("div",ae,[l(ut,{status:t.data.status},null,8,["status"]),e(),t.data.dataplaneType==="standard"?(i(),h(g,{key:0,items:t.data.dataplane.networking.inbounds,predicate:d=>d.state!=="Ready",empty:!1},{default:n(({items:d})=>[l(b,{class:"reason-tooltip"},{content:n(()=>[s("ul",null,[(i(!0),c(y,null,N(d,f=>(i(),c("li",{key:`${f.service}:${f.port}`},o(_("data-planes.routes.item.unhealthy_inbound",{service:f.service,port:f.port})),1))),128))])]),default:n(()=>[l(p(vt),{color:p(pt),size:p(J)},null,8,["color","size"]),e()]),_:2},1024)]),_:2},1032,["items","predicate"])):D("",!0)])]),_:2},1024),e(),K("use zones")&&t.data.zone?(i(),h(z,{key:0},{title:n(()=>[e(`
                    Zone
                  `)]),body:n(()=>[l(u,{to:{name:"zone-cp-detail-view",params:{zone:t.data.zone}}},{default:n(()=>[e(o(t.data.zone),1)]),_:1},8,["to"])]),_:1})):D("",!0),e(),l(z,null,{title:n(()=>[e(`
                    Type
                  `)]),body:n(()=>[e(o(_(`data-planes.type.${t.data.dataplaneType}`)),1)]),_:2},1024),e(),t.data.namespace.length>0?(i(),h(z,{key:1},{title:n(()=>[e(`
                    Namespace
                  `)]),body:n(()=>[e(o(t.data.namespace),1)]),_:1})):D("",!0)]),e(),s("div",ne,[l(z,null,{title:n(()=>[e(o(_("data-planes.routes.item.last_updated")),1)]),body:n(()=>[e(o(_("common.formats.datetime",{value:Date.parse(t.data.modificationTime)})),1)]),_:2},1024),e(),t.data.dataplane.networking.gateway?(i(),c(y,{key:0},[l(z,null,{title:n(()=>[e(o(_("http.api.property.tags")),1)]),body:n(()=>[l(ot,{tags:t.data.dataplane.networking.gateway.tags},null,8,["tags"])]),_:2},1024),e(),l(z,null,{title:n(()=>[e(o(_("http.api.property.address")),1)]),body:n(()=>[l(_t,{text:`${t.data.dataplane.networking.address}`},null,8,["text"])]),_:2},1024)],64)):D("",!0)])])]),_:2},1024),e(),l(w,{class:"traffic","data-testid":"dataplane-traffic"},{default:n(()=>[s("div",ie,[l(it,null,{title:n(()=>[l(p(kt),{display:"inline-block",decorative:"",size:p(J)},null,8,["size"]),e(`
                  Inbounds
                `)]),default:n(()=>[e(),l(W,{type:"inbound","data-testid":"dataplane-inbounds"},{default:n(()=>[(i(!0),c(y,null,N([t.data.dataplane.networking.type==="gateway"?Object.entries(($==null?void 0:$.inbounds)??{}).reduce((d,[f,m])=>{var q;const B=f.split("_").at(-1);return B===(((q=t.data.dataplane.networking.admin)==null?void 0:q.port)??"9901")?d:d.concat([{...t.data.dataplane.networking.inbounds[0],name:f,port:Number(B),protocol:["http","tcp"].find(H=>typeof m[H]<"u")??"tcp",addressPort:`${t.data.dataplane.networking.inbounds[0].address}:${B}`}])},[]):t.data.dataplane.networking.inbounds],d=>(i(),h(g,{key:d,items:d,predicate:f=>f.port!==49151},j({default:n(({items:f})=>[(i(!0),c(y,null,N(f,m=>(i(),c(y,{key:`${m.name}`},[(i(!0),c(y,null,N([$==null?void 0:$.inbounds[m.name]],B=>(i(),h(Q,{key:B,"data-testid":"dataplane-inbound",protocol:m.protocol,service:K("use service-insights",t.mesh)?m.tags["kuma.io/service"]:"",traffic:typeof E>"u"?B:{name:"",protocol:m.protocol,port:`${m.port}`}},{default:n(()=>[l(u,{"data-action":"",to:{name:(q=>q.includes("bound")?q.replace("-outbound-","-inbound-"):"connection-inbound-summary-overview-view")(String(p(a).name)),params:{connection:m.name},query:{inactive:x.params.inactive}}},{default:n(()=>[e(o(m.name.replace("localhost","").replace("_",":")),1)]),_:2},1032,["to"])]),_:2},1032,["protocol","service","traffic"]))),128))],64))),128))]),_:2},[t.data.dataplaneType==="delegated"?{name:"empty",fn:n(()=>[l(T,null,{default:n(()=>[s("p",null,`
                            This proxy is a delegated gateway therefore `+o(_("common.product.name"))+` does not have any visibility into inbounds for this gateway.
                          `,1)]),_:2},1024)]),key:"0"}:void 0]),1032,["items","predicate"]))),128))]),_:2},1024)]),_:2},1024),e(),l(it,null,j({title:n(()=>[l(p($t),{display:"inline-block",decorative:"",size:p(J)},null,8,["size"]),e(),se]),default:n(()=>[e(),e(),typeof E>"u"?(i(),c(y,{key:0},[typeof $>"u"?(i(),h(ft,{key:0})):(i(),c(y,{key:1},N(["upstream"],d=>(i(),c(y,{key:d},[l(W,{type:"passthrough"},{default:n(()=>[l(Q,{protocol:"passthrough",traffic:$.passthrough},{default:n(()=>[e(`
                          Non mesh traffic
                        `)]),_:2},1032,["traffic"])]),_:2},1024),e(),l(g,{predicate:x.params.inactive?void 0:([f,m])=>{var B,q;return((typeof m.tcp<"u"?(B=m.tcp)==null?void 0:B[`${d}_cx_rx_bytes_total`]:(q=m.http)==null?void 0:q[`${d}_rq_total`])??0)>0},items:Object.entries($.outbounds)},{default:n(({items:f})=>[f.length>0?(i(),h(W,{key:0,type:"outbound","data-testid":"dataplane-outbounds"},{default:n(()=>[(i(),c(y,null,N([/-([a-f0-9]){16}$/],m=>(i(),c(y,{key:m},[(i(!0),c(y,null,N(f,([B,q])=>(i(),h(Q,{key:`${B}`,"data-testid":"dataplane-outbound",protocol:["grpc","http","tcp"].find(H=>typeof q[H]<"u")??"tcp",traffic:q,service:q.$resourceMeta.type===""?B.replace(m,""):void 0,direction:d},{default:n(()=>[l(u,{"data-action":"",to:{name:(H=>H.includes("bound")?H.replace("-inbound-","-outbound-"):"connection-outbound-summary-overview-view")(String(p(a).name)),params:{connection:B},query:{inactive:x.params.inactive}}},{default:n(()=>[e(o(B),1)]),_:2},1032,["to"])]),_:2},1032,["protocol","traffic","service","direction"]))),128))],64))),64))]),_:2},1024)):D("",!0)]),_:2},1032,["predicate","items"])],64))),64))],64)):(i(),h(T,{key:1}))]),_:2},[$?{name:"actions",fn:n(()=>[l(S,{modelValue:x.params.inactive,"onUpdate:modelValue":d=>x.params.inactive=d,"data-testid":"dataplane-outbounds-inactive-toggle"},{label:n(()=>[e(`
                      Show inactive
                    `)]),_:2},1032,["modelValue","onUpdate:modelValue"]),e(),l(u,{action:"refresh",appearance:"primary",onClick:O},{default:n(()=>[e(`
                    Refresh
                  `)]),_:2},1032,["onClick"])]),key:"0"}:void 0]),1024)])]),_:2},1024),e(),l(V,null,{default:n(d=>[d.route.name!==x.name?(i(),h(ht,{key:0,width:"670px",onClose:function(){x.replace({name:"data-plane-detail-view",params:{mesh:x.params.mesh,dataPlane:x.params.dataPlane},query:{inactive:x.params.inactive?null:void 0}})}},{default:n(()=>[(i(),h(tt(d.Component),{data:x.params.subscription.length>0?t.data.dataplaneInsight.subscriptions:d.route.name.includes("-inbound-")?t.data.dataplane.networking.inbounds:($==null?void 0:$.outbounds)||{},"dataplane-overview":t.data},null,8,["data","dataplane-overview"]))]),_:2},1032,["onClose"])):D("",!0)]),_:2},1024),e(),s("div",oe,[s("h2",null,o(_("data-planes.routes.item.mtls.title")),1),e(),t.data.dataplaneInsight.mTLS?(i(!0),c(y,{key:0},N([t.data.dataplaneInsight.mTLS],d=>(i(),h(w,{key:d,class:"mt-4"},{default:n(()=>[s("div",re,[l(z,null,{title:n(()=>[e(o(_("data-planes.routes.item.mtls.expiration_time.title")),1)]),body:n(()=>[e(o(_("common.formats.datetime",{value:Date.parse(d.certificateExpirationTime)})),1)]),_:2},1024),e(),l(z,null,{title:n(()=>[e(o(_("data-planes.routes.item.mtls.generation_time.title")),1)]),body:n(()=>[e(o(_("common.formats.datetime",{value:Date.parse(d.lastCertificateRegeneration)})),1)]),_:2},1024),e(),l(z,null,{title:n(()=>[e(o(_("data-planes.routes.item.mtls.regenerations.title")),1)]),body:n(()=>[e(o(_("common.formats.integer",{value:d.certificateRegenerations})),1)]),_:2},1024),e(),l(z,null,{title:n(()=>[e(o(_("data-planes.routes.item.mtls.issued_backend.title")),1)]),body:n(()=>[e(o(d.issuedBackend),1)]),_:2},1024),e(),l(z,null,{title:n(()=>[e(o(_("data-planes.routes.item.mtls.supported_backends.title")),1)]),body:n(()=>[s("ul",null,[(i(!0),c(y,null,N(d.supportedBackends,f=>(i(),c("li",{key:f},o(f),1))),128))])]),_:2},1024)])]),_:2},1024))),128)):(i(),h(U,{key:1,class:"mt-4",appearance:"warning"},{default:n(()=>[s("div",{innerHTML:_("data-planes.routes.item.mtls.disabled")},null,8,de)]),_:2},1024))]),e(),t.data.dataplaneInsight.subscriptions.length>0?(i(),c("div",le,[s("h2",null,o(_("data-planes.routes.item.subscriptions.title")),1),e(),l(mt,{headers:[{...v.get("headers.instanceId"),label:_("http.api.property.instanceId"),key:"instanceId"},{...v.get("headers.version"),label:_("http.api.property.version"),key:"version"},{...v.get("headers.connected"),label:_("http.api.property.connected"),key:"connected"},{...v.get("headers.disconnected"),label:_("http.api.property.disconnected"),key:"disconnected"},{...v.get("headers.responses"),label:_("http.api.property.responses"),key:"responses"}],"is-selected-row":d=>d.id===x.params.subscription,items:t.data.dataplaneInsight.subscriptions.map((d,f,m)=>m[m.length-(f+1)]),onResize:v.set},{instanceId:n(({row:d})=>[l(u,{"data-action":"",to:{name:"data-plane-subscription-summary-view",params:{subscription:d.id}}},{default:n(()=>[e(o(d.controlPlaneInstanceId),1)]),_:2},1032,["to"])]),version:n(({row:d})=>{var f,m;return[e(o(((m=(f=d.version)==null?void 0:f.kumaDp)==null?void 0:m.version)??"-"),1)]}),connected:n(({row:d})=>[e(o(_("common.formats.datetime",{value:Date.parse(d.connectTime??"")})),1)]),disconnected:n(({row:d})=>[d.disconnectTime?(i(),c(y,{key:0},[e(o(_("common.formats.datetime",{value:Date.parse(d.disconnectTime)})),1)],64)):D("",!0)]),responses:n(({row:d})=>{var f;return[(i(!0),c(y,null,N([((f=d.status)==null?void 0:f.total)??{}],m=>(i(),c(y,null,[e(o(m.responsesSent)+"/"+o(m.responsesAcknowledged),1)],64))),256))]}),_:2},1032,["headers","is-selected-row","items","onResize"])])):D("",!0)])]),_:2},[k.value.length>0||E?{name:"notifications",fn:n(()=>[s("ul",jt,[(i(!0),c(y,null,N(k.value,d=>(i(),c("li",{key:d.kind,"data-testid":`warning-${d.kind}`,innerHTML:_(`common.warnings.${d.kind}`,d.payload)},null,8,Jt))),128)),e(),E?(i(),c("li",Qt,[e(`
              The below view is not enhanced with runtime stats (Error loading stats: `),s("strong",null,o(E.toString()),1),e(`)
            `)])):D("",!0)])]),key:"0"}:void 0]),1024)]),_:2},1032,["src"])]),_:1})}}}),me=A(ce,[["__scopeId","data-v-5bed4307"]]);export{me as default};
