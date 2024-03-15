import{_ as M,a as k,o as s,b as g,w as a,r as N,m as n,f as t,d as G,l as tt,p as d,t as i,e as l,F as f,c as u,H as w,q as K,n as et,B as at,M as nt,Z as T,$ as ot,a0 as st,K as z,W as it,a1 as lt,a2 as W,X as O,a3 as rt,A as dt,a4 as ct,D as ut,x as pt,y as _t}from"./index-291i1Sgi.js";import{S as mt}from"./StatusBadge-fkLFYfX6.js";import{S as ft}from"./SummaryView-U6wIlsXk.js";import{T as X}from"./TagList-o31LB9kK.js";import{_ as yt}from"./SubscriptionList.vue_vue_type_script_setup_true_lang-je8TUnPu.js";import"./AccordionList-zqgqeSCG.js";const vt=["B","kB","MB","GB","TB","PB","EB","ZB","YB"],gt=["B","KiB","MiB","GiB","TiB","PiB","EiB","ZiB","YiB"],ht=["b","kbit","Mbit","Gbit","Tbit","Pbit","Ebit","Zbit","Ybit"],kt=["b","kibit","Mibit","Gibit","Tibit","Pibit","Eibit","Zibit","Yibit"],j=(c,o,e)=>{let b=c;return typeof o=="string"||Array.isArray(o)?b=c.toLocaleString(o,e):(o===!0||e!==void 0)&&(b=c.toLocaleString(void 0,e)),b};function F(c,o){if(!Number.isFinite(c))throw new TypeError(`Expected a finite number, got ${typeof c}: ${c}`);o={bits:!1,binary:!1,space:!0,...o};const e=o.bits?o.binary?kt:ht:o.binary?gt:vt,b=o.space?" ":"";if(o.signed&&c===0)return` 0${b}${e[0]}`;const x=c<0,S=x?"-":o.signed?"+":"";x&&(c=-c);let h;if(o.minimumFractionDigits!==void 0&&(h={minimumFractionDigits:o.minimumFractionDigits}),o.maximumFractionDigits!==void 0&&(h={maximumFractionDigits:o.maximumFractionDigits,...h}),c<1){const E=j(c,o.locale,h);return S+E+b+e[0]}const $=Math.min(Math.floor(o.binary?Math.log(c)/Math.log(1024):Math.log10(c)/3),e.length-1);c/=(o.binary?1024:1e3)**$,h||(c=c.toPrecision(3));const B=j(Number(c),o.locale,h),D=e[$];return S+B+b+D}const bt={},$t={class:"card"},wt={class:"title"},xt={class:"body"};function Bt(c,o){const e=k("KCard");return s(),g(e,{class:"data-card"},{default:a(()=>[n("dl",null,[n("div",$t,[n("dt",wt,[N(c.$slots,"title",{},void 0,!0)]),t(),n("dd",xt,[N(c.$slots,"default",{},void 0,!0)])])])]),_:3})}const Q=M(bt,[["render",Bt],["__scopeId","data-v-6e083223"]]),Ct={class:"title"},Tt={key:0},It=G({__name:"ConnectionCard",props:{protocol:{},traffic:{default:void 0},direction:{default:"downstream"}},setup(c){const{t:o}=tt(),e=c,b=x=>{const S=x.target;if(x.isTrusted&&S.nodeName.toLowerCase()!=="a"){const h=S.closest(".service-traffic-card, a");if(h){const $=h.nodeName.toLowerCase()==="a"?h:h.querySelector("a");$!==null&&"click"in $&&typeof $.click=="function"&&$.click()}}};return(x,S)=>{const h=k("KBadge"),$=k("KSkeletonBox");return s(),g(Q,{class:"service-traffic-card",onClick:b},{title:a(()=>[l(h,{appearance:e.protocol==="passthrough"?"success":"info"},{default:a(()=>[t(i(d(o)(`data-planes.components.service_traffic_card.protocol.${e.protocol}`,{},{defaultMessage:d(o)(`http.api.value.${e.protocol}`)})),1)]),_:1},8,["appearance"]),t(),n("div",Ct,[N(x.$slots,"default",{},void 0,!0)])]),default:a(()=>{var B,D,E,R,L,P,q,U,A;return[t(),e.traffic?(s(),u(f,{key:0},[e.traffic.name.length>0?(s(),u("dl",Tt,[e.protocol==="passthrough"?(s(!0),u(f,{key:0},w([["http","tcp"].reduce((p,m)=>{var I;const y=e.direction;return Object.entries(((I=e.traffic)==null?void 0:I[m])||{}).reduce((V,[r,_])=>[`${y}_cx_tx_bytes_total`,`${y}_cx_rx_bytes_total`].includes(r)?{...V,[r]:_+(V[r]??0)}:V,p)},{})],(p,m)=>(s(),u(f,{key:m},[n("div",null,[n("dt",null,i(d(o)("data-planes.components.service_traffic_card.tx")),1),t(),n("dd",null,i(d(F)(p.downstream_cx_rx_bytes_total??0)),1)]),t(),n("div",null,[n("dt",null,i(d(o)("data-planes.components.service_traffic_card.rx")),1),t(),n("dd",null,i(d(F)(p.downstream_cx_tx_bytes_total??0)),1)])],64))),128)):e.protocol==="grpc"?(s(),u(f,{key:1},[n("div",null,[n("dt",null,i(d(o)("data-planes.components.service_traffic_card.grpc_success")),1),t(),n("dd",null,i(d(o)("common.formats.integer",{value:((B=e.traffic.grpc)==null?void 0:B.success)??0})),1)]),t(),n("div",null,[n("dt",null,i(d(o)("data-planes.components.service_traffic_card.grpc_failure")),1),t(),n("dd",null,i(d(o)("common.formats.integer",{value:((D=e.traffic.grpc)==null?void 0:D.failure)??0})),1)])],64)):e.protocol.startsWith("http")?(s(),u(f,{key:2},[(s(!0),u(f,null,w([((E=e.traffic.http)==null?void 0:E[`${e.direction}_rq_1xx`])??0].filter(p=>p!==0),p=>(s(),u("div",{key:p},[n("dt",null,i(d(o)("data-planes.components.service_traffic_card.1xx")),1),t(),n("dd",null,i(d(o)("common.formats.integer",{value:p})),1)]))),128)),t(),n("div",null,[n("dt",null,i(d(o)("data-planes.components.service_traffic_card.2xx")),1),t(),n("dd",null,i(d(o)("common.formats.integer",{value:((R=e.traffic.http)==null?void 0:R[`${e.direction}_rq_2xx`])??0})),1)]),t(),(s(!0),u(f,null,w([((L=e.traffic.http)==null?void 0:L[`${e.direction}_rq_3xx`])??0].filter(p=>p!==0),p=>(s(),u("div",{key:p},[n("dt",null,i(d(o)("data-planes.components.service_traffic_card.3xx")),1),t(),n("dd",null,i(d(o)("common.formats.integer",{value:p})),1)]))),128)),t(),n("div",null,[n("dt",null,i(d(o)("data-planes.components.service_traffic_card.4xx")),1),t(),n("dd",null,i(d(o)("common.formats.integer",{value:((P=e.traffic.http)==null?void 0:P[`${e.direction}_rq_4xx`])??0})),1)]),t(),n("div",null,[n("dt",null,i(d(o)("data-planes.components.service_traffic_card.5xx")),1),t(),n("dd",null,i(d(o)("common.formats.integer",{value:((q=e.traffic.http)==null?void 0:q[`${e.direction}_rq_5xx`])??0})),1)])],64)):(s(),u(f,{key:3},[n("div",null,[n("dt",null,i(d(o)("data-planes.components.service_traffic_card.tx")),1),t(),n("dd",null,i(d(F)(((U=e.traffic.tcp)==null?void 0:U[`${e.direction}_cx_rx_bytes_total`])??0)),1)]),t(),n("div",null,[n("dt",null,i(d(o)("data-planes.components.service_traffic_card.rx")),1),t(),n("dd",null,i(d(F)(((A=e.traffic.tcp)==null?void 0:A[`${e.direction}_cx_tx_bytes_total`])??0)),1)])],64))])):K("",!0)],64)):(s(),g($,{key:1,width:"10"}))]}),_:3})}}}),Y=M(It,[["__scopeId","data-v-fc727eed"]]),St={class:"body"},Dt=G({__name:"ConnectionGroup",props:{type:{}},setup(c){const o=c;return(e,b)=>{const x=k("KCard");return s(),g(x,{class:et(["service-traffic-group",`type-${o.type}`])},{default:a(()=>[n("div",St,[N(e.$slots,"default",{},void 0,!0)])]),_:3},8,["class"])}}}),Z=M(Dt,[["__scopeId","data-v-9402d5d1"]]),Kt={class:"service-traffic"},Nt={class:"actions"},Et=G({__name:"ConnectionTraffic",setup(c){return(o,e)=>(s(),u("div",Kt,[n("div",Nt,[N(o.$slots,"actions",{},void 0,!0)]),t(),l(Q,{class:"header"},{title:a(()=>[N(o.$slots,"title",{},void 0,!0)]),_:3}),t(),N(o.$slots,"default",{},void 0,!0)]))}}),J=M(Et,[["__scopeId","data-v-e6bd176c"]]),Vt=c=>(pt("data-v-fcf6f62d"),c=c(),_t(),c),Mt={"data-testid":"dataplane-warnings"},Rt=["data-testid","innerHTML"],Lt={key:0,"data-testid":"warning-stats-loading"},Pt={class:"stack","data-testid":"dataplane-details"},qt={class:"columns"},Ut={class:"status-with-reason"},At={class:"columns"},Ft=Vt(()=>n("span",null,"Outbounds",-1)),Gt={"data-testid":"dataplane-mtls"},zt={class:"columns"},Ot=["innerHTML"],Yt={key:0,"data-testid":"dataplane-subscriptions"},Zt=G({__name:"DataPlaneDetailView",props:{data:{}},setup(c){const o=at(),e=c,b=nt(()=>e.data.warnings.concat(...e.data.isCertExpired?[{kind:"CERT_EXPIRED"}]:[]));return(x,S)=>{const h=k("KTooltip"),$=k("DataCollection"),B=k("KCard"),D=k("RouterLink"),E=k("KInputSwitch"),R=k("KButton"),L=k("RouterView"),P=k("KAlert"),q=k("AppView"),U=k("DataSource"),A=k("RouteView");return s(),g(A,{params:{mesh:"",dataPlane:"",inactive:!1},name:"data-plane-detail-view"},{default:a(({route:p,t:m})=>[l(U,{src:`/meshes/${p.params.mesh}/dataplanes/${p.params.dataPlane}/stats/${e.data.dataplane.networking.inboundName}`},{default:a(({data:y,error:I,refresh:V})=>[l(q,null,O({default:a(()=>[t(),n("div",Pt,[l(B,null,{default:a(()=>[n("div",qt,[l(T,null,{title:a(()=>[t(i(m("http.api.property.status")),1)]),body:a(()=>[n("div",Ut,[l(mt,{status:e.data.status},null,8,["status"]),t(),e.data.dataplane.networking.type==="standard"?(s(),g($,{key:0,items:e.data.dataplane.networking.inbounds,predicate:r=>!r.health.ready,empty:!1},{default:a(({items:r})=>[l(h,{class:"reason-tooltip"},{content:a(()=>[n("ul",null,[(s(!0),u(f,null,w(r,_=>(s(),u("li",{key:`${_.service}:${_.port}`},i(m("data-planes.routes.item.unhealthy_inbound",{service:_.service,port:_.port})),1))),128))])]),default:a(()=>[l(d(ot),{color:d(st),size:d(z)},null,8,["color","size"]),t()]),_:2},1024)]),_:2},1032,["items","predicate"])):K("",!0)])]),_:2},1024),t(),l(T,null,{title:a(()=>[t(i(m("data-planes.routes.item.last_updated")),1)]),body:a(()=>[t(i(m("common.formats.datetime",{value:Date.parse(e.data.modificationTime)})),1)]),_:2},1024),t(),e.data.dataplane.networking.gateway?(s(),u(f,{key:0},[l(T,null,{title:a(()=>[t(i(m("http.api.property.tags")),1)]),body:a(()=>[l(X,{tags:e.data.dataplane.networking.gateway.tags},null,8,["tags"])]),_:2},1024),t(),l(T,null,{title:a(()=>[t(i(m("http.api.property.address")),1)]),body:a(()=>[l(it,{text:`${e.data.dataplane.networking.address}`},null,8,["text"])]),_:2},1024)],64)):K("",!0)])]),_:2},1024),t(),l(B,{class:"traffic","data-testid":"dataplane-traffic"},{default:a(()=>[n("div",At,[l(J,null,{title:a(()=>[l(d(lt),{decorative:"",size:d(z)},null,8,["size"]),t(`
                  Inbounds
                `)]),default:a(()=>[t(),l(Z,{type:"inbound"},{default:a(()=>[l($,{items:e.data.dataplane.networking.gateway?((y==null?void 0:y.inbounds)??[]).map(r=>({...e.data.dataplane.networking.inbounds[0],name:r.name,port:Number(r.port),protocol:r.protocol})):e.data.dataplane.networking.inbounds},O({default:a(({items:r})=>[(s(!0),u(f,null,w(r,_=>(s(),u(f,{key:`${_.name}`},[(s(!0),u(f,null,w([(y||{inbounds:[]}).inbounds.find(v=>`${v.name}`==`${_.name}`)],v=>(s(),g(Y,{key:v,protocol:(v==null?void 0:v.protocol)??_.protocol,traffic:typeof I>"u"?v:{name:"",protocol:_.protocol,port:`${_.port}`}},{default:a(()=>[(s(!0),u(f,null,w([`${_.name.replace(`_${_.port}`,"").replace("localhost","")}:${_.port}`],C=>(s(),g(D,{key:C,to:{name:(H=>H.includes("bound")?H.replace("-outbound-","-inbound-"):"connection-inbound-summary-overview-view")(String(d(o).name)),params:{service:C},query:{inactive:p.params.inactive?null:void 0}}},{default:a(()=>[t(i(C),1)]),_:2},1032,["to"]))),128)),t(),l(X,{tags:[{label:"kuma.io/service",value:_.tags["kuma.io/service"]}]},null,8,["tags"])]),_:2},1032,["protocol","traffic"]))),128))],64))),128))]),_:2},[e.data.dataplaneType==="delegated"?{name:"empty",fn:a(()=>[l(W,null,{default:a(()=>[n("p",null,"This proxy is a delegated gateway therefore "+i(m("common.product.name"))+" does not have any visibility into inbounds for this gateway",1)]),_:2},1024)]),key:"0"}:void 0]),1032,["items"])]),_:2},1024)]),_:2},1024),t(),l(J,null,O({title:a(()=>[l(d(rt),{decorative:"",size:d(z)},null,8,["size"]),t(),Ft]),default:a(()=>[t(),t(),typeof I>"u"?(s(),u(f,{key:0},[typeof y>"u"?(s(),g(dt,{key:0})):(s(),u(f,{key:1},w(["upstream"],r=>(s(),u(f,{key:r},[l(Z,{type:"passthrough"},{default:a(()=>[l(Y,{protocol:"passthrough",traffic:y.passthrough},{default:a(()=>[t(`
                          Non mesh traffic
                        `)]),_:2},1032,["traffic"])]),_:2},1024),t(),l($,{predicate:p.params.inactive?void 0:_=>{var v,C;return((_.protocol==="tcp"?(v=_.tcp)==null?void 0:v[`${r}_cx_rx_bytes_total`]:(C=_.http)==null?void 0:C[`${r}_rq_total`])??0)>0},items:y.outbounds},{default:a(({items:_})=>[_.length>0?(s(),g(Z,{key:0,type:"outbound","data-testid":"dataplane-outbounds"},{default:a(()=>[(s(!0),u(f,null,w(_,v=>(s(),g(Y,{key:`${v.name}`,protocol:v.protocol,traffic:v,direction:r},{default:a(()=>[l(D,{to:{name:(C=>C.includes("bound")?C.replace("-inbound-","-outbound-"):"connection-outbound-summary-overview-view")(String(d(o).name)),params:{service:v.name},query:{inactive:p.params.inactive?null:void 0}}},{default:a(()=>[t(i(v.name),1)]),_:2},1032,["to"])]),_:2},1032,["protocol","traffic","direction"]))),128))]),_:2},1024)):K("",!0)]),_:2},1032,["predicate","items"])],64))),64))],64)):(s(),g(W,{key:1}))]),_:2},[y?{name:"actions",fn:a(()=>[l(E,{modelValue:p.params.inactive,"onUpdate:modelValue":r=>p.params.inactive=r,"data-testid":"dataplane-outbounds-inactive-toggle"},{label:a(()=>[t(`
                      Show inactive
                    `)]),_:2},1032,["modelValue","onUpdate:modelValue"]),t(),l(R,{appearance:"primary",onClick:V},{default:a(()=>[l(d(ct)),t(`

                    Refresh
                  `)]),_:2},1032,["onClick"])]),key:"0"}:void 0]),1024)])]),_:2},1024),t(),l(L,null,{default:a(r=>[r.route.name!==p.name?(s(),g(ft,{key:0,width:"670px",onClose:function(_){p.replace({name:"data-plane-detail-view",params:{mesh:p.params.mesh,dataPlane:p.params.dataPlane},query:{inactive:p.params.inactive?null:void 0}})}},{default:a(()=>[(s(),g(ut(r.Component),{"dataplane-type":e.data.dataplaneType,gateway:e.data.dataplane.networking.gateway,inbounds:r.route.name.includes("-inbound-")?e.data.dataplane.networking.inbounds:[],data:r.route.name.includes("-inbound-")?(y==null?void 0:y.inbounds)||[]:(y==null?void 0:y.outbounds)||[]},null,8,["dataplane-type","gateway","inbounds","data"]))]),_:2},1032,["onClose"])):K("",!0)]),_:2},1024),t(),n("div",Gt,[n("h2",null,i(m("data-planes.routes.item.mtls.title")),1),t(),e.data.dataplaneInsight.mTLS?(s(!0),u(f,{key:0},w([e.data.dataplaneInsight.mTLS],r=>(s(),g(B,{key:r,class:"mt-4"},{default:a(()=>[n("div",zt,[l(T,null,{title:a(()=>[t(i(m("data-planes.routes.item.mtls.expiration_time.title")),1)]),body:a(()=>[t(i(m("common.formats.datetime",{value:Date.parse(r.certificateExpirationTime)})),1)]),_:2},1024),t(),l(T,null,{title:a(()=>[t(i(m("data-planes.routes.item.mtls.generation_time.title")),1)]),body:a(()=>[t(i(m("common.formats.datetime",{value:Date.parse(r.lastCertificateRegeneration)})),1)]),_:2},1024),t(),l(T,null,{title:a(()=>[t(i(m("data-planes.routes.item.mtls.regenerations.title")),1)]),body:a(()=>[t(i(m("common.formats.integer",{value:r.certificateRegenerations})),1)]),_:2},1024),t(),l(T,null,{title:a(()=>[t(i(m("data-planes.routes.item.mtls.issued_backend.title")),1)]),body:a(()=>[t(i(r.issuedBackend),1)]),_:2},1024),t(),l(T,null,{title:a(()=>[t(i(m("data-planes.routes.item.mtls.supported_backends.title")),1)]),body:a(()=>[n("ul",null,[(s(!0),u(f,null,w(r.supportedBackends,_=>(s(),u("li",{key:_},i(_),1))),128))])]),_:2},1024)])]),_:2},1024))),128)):(s(),g(P,{key:1,class:"mt-4",appearance:"warning"},{default:a(()=>[n("div",{innerHTML:m("data-planes.routes.item.mtls.disabled")},null,8,Ot)]),_:2},1024))]),t(),e.data.dataplaneInsight.subscriptions.length>0?(s(),u("div",Yt,[n("h2",null,i(m("data-planes.routes.item.subscriptions.title")),1),t(),l(B,{class:"mt-4"},{default:a(()=>[l(yt,{subscriptions:e.data.dataplaneInsight.subscriptions},null,8,["subscriptions"])]),_:1})])):K("",!0)])]),_:2},[b.value.length>0||I?{name:"notifications",fn:a(()=>[n("ul",Mt,[(s(!0),u(f,null,w(b.value,r=>(s(),u("li",{key:r.kind,"data-testid":`warning-${r.kind}`,innerHTML:m(`common.warnings.${r.kind}`,r.payload)},null,8,Rt))),128)),t(),I?(s(),u("li",Lt,[t(`
              The below view is not enhanced with runtime stats (Error loading stats: `),n("strong",null,i(I.toString()),1),t(`)
            `)])):K("",!0)])]),key:"0"}:void 0]),1024)]),_:2},1032,["src"])]),_:1})}}}),te=M(Zt,[["__scopeId","data-v-fcf6f62d"]]);export{te as default};
