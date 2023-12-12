import{a as X,K as P}from"./index-fce48c05.js";import{_ as D,a as g,o as r,b as f,w as e,p as o,r as w,f as t,d as q,c,e as l,l as U,t as n,q as a,s as $,n as Z,M as J,a0 as Q,a1 as y,a2 as Y,F as m,a3 as tt,I as x,a4 as et,y as at,z as st}from"./index-287cffdc.js";import{E as ot}from"./ErrorBlock-d2b1e8cd.js";import{_ as nt}from"./LoadingBlock.vue_vue_type_script_setup_true_lang-69adf9f8.js";import{S as lt}from"./StatusBadge-9d53f593.js";import{T as A}from"./TagList-c8b87d31.js";import{T as K}from"./TextWithCopyButton-270d741f.js";import{_ as rt}from"./SubscriptionList.vue_vue_type_script_setup_true_lang-8703c5d4.js";import"./WarningIcon.vue_vue_type_script_setup_true_lang-0e3545a8.js";import"./CopyButton-c63b5581.js";import"./AccordionList-4de6d375.js";const dt={},it={class:"card"},ct={class:"title"},ut={class:"body"};function _t(h,s){const p=g("KCard");return r(),f(p,{class:"data-card"},{default:e(()=>[o("dl",null,[o("div",it,[o("dt",ct,[w(h.$slots,"title",{},void 0,!0)]),t(),o("dd",ut,[w(h.$slots,"default",{},void 0,!0)])])])]),_:3})}const L=D(dt,[["render",_t],["__scopeId","data-v-65e18720"]]),pt={class:"service-traffic"},ft=q({__name:"DataPlaneTraffic",setup(h){return(s,p)=>(r(),c("div",pt,[l(L,{class:"header"},{title:e(()=>[w(s.$slots,"title",{},void 0,!0)]),_:3}),t(),w(s.$slots,"default",{},void 0,!0)]))}});const N=D(ft,[["__scopeId","data-v-8e8bcc45"]]),mt={class:"title"},ht={key:0},yt=q({__name:"ServiceTrafficCard",props:{protocol:{},requests:{default:void 0},rx:{default:0},tx:{default:0}},setup(h){const{t:s}=U(),p=h;return(u,C)=>{const V=g("KBadge");return r(),f(L,{class:"service-traffic-card"},{title:e(()=>[l(V,{appearance:p.protocol==="unknown"?"success":"info"},{default:e(()=>[t(n(a(s)(`data-planes.components.service_traffic_card.protocol.${p.protocol}`,{},{defaultMessage:a(s)(`http.api.value.${p.protocol}`)})),1)]),_:1},8,["appearance"]),t(),o("span",mt,[w(u.$slots,"default",{},void 0,!0)])]),default:e(()=>[t(),o("dl",null,[o("div",null,[o("dt",null,n(a(s)("data-planes.components.service_traffic_card.tx")),1),t(),o("dd",null,n(a(s)("common.formats.integer",{value:p.tx})),1)]),t(),o("div",null,[o("dt",null,n(a(s)("data-planes.components.service_traffic_card.rx")),1),t(),o("dd",null,n(a(s)("common.formats.integer",{value:p.rx})),1)]),t(),typeof p.requests<"u"?(r(),c("div",ht,[o("dt",null,n(a(s)("data-planes.components.service_traffic_card.requests")),1),t(),o("dd",null,n(a(s)("common.formats.integer",{value:p.requests})),1)])):$("",!0)])]),_:3})}}});const R=D(yt,[["__scopeId","data-v-19bc134f"]]),vt={class:"body"},gt=q({__name:"ServiceTrafficGroup",props:{type:{}},setup(h){const s=h;return(p,u)=>{const C=g("KCard");return r(),f(C,{class:Z(["service-traffic-group",`type-${s.type}`])},{default:e(()=>[o("div",vt,[w(p.$slots,"default",{},void 0,!0)])]),_:3},8,["class"])}}});const M=D(gt,[["__scopeId","data-v-baf4abf7"]]),bt=h=>(at("data-v-f8375696"),h=h(),st(),h),xt={"data-testid":"dataplane-warnings"},kt=["data-testid","innerHTML"],$t={class:"stack","data-testid":"dataplane-details"},wt={class:"columns"},Ct={class:"status-with-reason"},It={class:"columns"},Tt=bt(()=>o("span",null,"Outbounds",-1)),Dt={class:"passthrough"},Kt={class:"outbounds"},qt={key:1,"data-testid":"dataplane-inbounds"},Bt={class:"inbound-list"},St={class:"mt-4 columns"},Pt={"data-testid":"dataplane-mtls"},Rt={class:"columns"},Mt=["innerHTML"],Vt={key:2,"data-testid":"dataplane-subscriptions"},Et=q({__name:"DataPlaneDetailView",props:{data:{}},setup(h){const{t:s,formatIsoDate:p}=U(),u=h,C=J(()=>u.data.warnings.concat(...u.data.isCertExpired?[{kind:"CERT_EXPIRED"}]:[]));return(V,zt)=>{const O=g("KTooltip"),I=g("KCard"),G=g("DataSource"),E=g("KBadge"),H=g("KAlert"),F=g("AppView"),j=g("RouteView");return r(),f(j,{params:{mesh:"",dataPlane:""},name:"data-plane-detail-view"},{default:e(({can:W,route:z})=>[l(F,null,Q({default:e(()=>[t(),o("div",$t,[l(I,null,{default:e(()=>[o("div",wt,[l(y,null,{title:e(()=>[t(n(a(s)("http.api.property.status")),1)]),body:e(()=>[o("div",Ct,[l(lt,{status:u.data.status},null,8,["status"]),t(),u.data.unhealthyInbounds.length>0?(r(),f(O,{key:0,label:u.data.unhealthyInbounds.map(d=>a(s)("data-planes.routes.item.unhealthy_inbound",d)).join(", "),class:"reason-tooltip"},{default:e(()=>[l(a(Y),{color:a(X),size:a(P),"hide-title":""},null,8,["color","size"])]),_:1},8,["label"])):$("",!0)])]),_:1}),t(),l(y,null,{title:e(()=>[t(n(a(s)("data-planes.routes.item.last_updated")),1)]),body:e(()=>[u.data.lastUpdateTime?(r(),c(m,{key:0},[t(n(a(p)(u.data.lastUpdateTime)),1)],64)):(r(),c(m,{key:1},[t(n(a(s)("common.detail.none")),1)],64))]),_:1}),t(),u.data.dataplane.networking.gateway?(r(),c(m,{key:0},[l(y,null,{title:e(()=>[t(n(a(s)("http.api.property.tags")),1)]),body:e(()=>[l(A,{tags:u.data.dataplane.networking.gateway.tags},null,8,["tags"])]),_:1}),t(),l(y,null,{title:e(()=>[t(n(a(s)("http.api.property.address")),1)]),body:e(()=>[l(K,{text:`${u.data.dataplane.networking.address}`},null,8,["text"])]),_:1})],64)):$("",!0)])]),_:1}),t(),W("read traffic")&&u.data.dataplaneType==="standard"?(r(),f(G,{key:0,src:`/meshes/${z.params.mesh}/dataplanes/${z.params.dataPlane}/traffic`},{default:e(({data:d,error:k})=>[k?(r(),f(ot,{key:0,error:k},null,8,["error"])):d===void 0?(r(),f(nt,{key:1})):(r(),f(I,{key:2,class:"traffic"},{default:e(()=>[o("div",It,[l(N,null,{title:e(()=>[l(a(tt),{display:"inline-block",decorative:"",size:a(P)},null,8,["size"]),t(`
                  Inbounds
                `)]),data:e(()=>[o("dl",null,[o("div",null,[o("dt",null,n(a(s)("services.components.service_traffic.inbound",{},{defaultMessage:"Requests"})),1),t(),o("dd",null,n(a(s)("common.formats.integer",{value:1e3})),1)])])]),default:e(()=>[t(),t(),l(M,{type:"inbound"},{default:e(()=>[(r(!0),c(m,null,x(d.inbounds,_=>(r(),c(m,{key:`${_.name}`},[(r(!0),c(m,null,x([{protocol:typeof _.http<"u"?"http":"tcp",direction:"downstream"}],i=>{var b,v;return r(),f(R,{key:i.protocol,protocol:i.protocol,tx:(b=_[i.protocol])==null?void 0:b[`${i.direction}_cx_tx_bytes_total`],rx:(v=_[i.protocol])==null?void 0:v[`${i.direction}_cx_rx_bytes_total`],requests:i.protocol==="http"?["http1_total","http2_total","http3_total"].reduce((B,S)=>{var T;return B+(((T=_.http)==null?void 0:T[`${i.direction}_rq_${S}`])??0)},0):void 0},{default:e(()=>[t(n(_.name),1)]),_:2},1032,["protocol","tx","rx","requests"])}),128))],64))),128))]),_:2},1024)]),_:2},1024),t(),l(N,null,{title:e(()=>[l(a(et),{display:"inline-block",decorative:"",size:a(P)},null,8,["size"]),t(),Tt]),data:e(()=>[o("dl",null,[o("div",null,[o("dt",Dt,n(a(s)("services.components.service_traffic.passthrough",{},{defaultMessage:"Passthrough Requests"})),1),t(),o("dd",null,n(a(s)("common.formats.integer",{value:1e3})),1)]),t(),o("div",null,[o("dt",Kt,n(a(s)("services.components.service_traffic.mesh",{},{defaultMessage:"Mesh Requests"})),1),t(),o("dd",null,n(a(s)("common.formats.integer",{value:1e3})),1)])])]),default:e(()=>[t(),t(),l(M,{type:"passthrough"},{default:e(()=>[(r(),c(m,null,x([{protocol:"cluster",direction:"downstream"}],_=>l(R,{key:_.protocol,protocol:"unknown",tx:d.passthrough.reduce((i,b)=>{var v;return i+(((v=b[_.protocol])==null?void 0:v[`${_.direction}_cx_tx_bytes_total`])??0)},0),rx:d.passthrough.reduce((i,b)=>{var v;return i+(((v=b[_.protocol])==null?void 0:v[`${_.direction}_cx_rx_bytes_total`])??0)},0)},{default:e(()=>[t(`
                      Non mesh traffic
                    `)]),_:2},1032,["tx","rx"])),64))]),_:2},1024),t(),l(M,{type:"outbound"},{default:e(()=>[(r(!0),c(m,null,x(d.outbounds,_=>(r(),c(m,{key:`${_.name}`},[(r(!0),c(m,null,x([{protocol:typeof _.http<"u"?"http":"tcp",direction:"downstream"}],i=>{var b,v;return r(),f(R,{key:i.protocol,protocol:i.protocol,tx:(b=_[i.protocol])==null?void 0:b[`${i.direction}_cx_tx_bytes_total`],rx:(v=_[i.protocol])==null?void 0:v[`${i.direction}_cx_rx_bytes_total`],requests:i.protocol==="http"?["http1_total","http2_total","http3_total"].reduce((B,S)=>{var T;return B+(((T=_.http)==null?void 0:T[`${i.direction}_rq_${S}`])??0)},0):void 0},{default:e(()=>[t(n(_.name),1)]),_:2},1032,["protocol","tx","rx","requests"])}),128))],64))),128))]),_:2},1024)]),_:2},1024)])]),_:2},1024))]),_:2},1032,["src"])):$("",!0),t(),u.data.dataplane.networking.inbounds.length>0?(r(),c("div",qt,[o("h2",null,n(a(s)("data-planes.routes.item.inbounds")),1),t(),l(I,{class:"mt-4"},{default:e(()=>[o("div",Bt,[(r(!0),c(m,null,x(u.data.dataplane.networking.inbounds,(d,k)=>(r(),c("div",{key:k,class:"inbound"},[o("h4",null,[l(K,{text:d.tags["kuma.io/service"]},{default:e(()=>[t(n(a(s)("data-planes.routes.item.inbound_name",{service:d.tags["kuma.io/service"]})),1)]),_:2},1032,["text"])]),t(),o("div",St,[l(y,null,{title:e(()=>[t(n(a(s)("http.api.property.status")),1)]),body:e(()=>[d.health.ready?(r(),f(E,{key:0,appearance:"success"},{default:e(()=>[t(n(a(s)("data-planes.routes.item.health.ready")),1)]),_:1})):(r(),f(E,{key:1,appearance:"danger"},{default:e(()=>[t(n(a(s)("data-planes.routes.item.health.not_ready")),1)]),_:1}))]),_:2},1024),t(),l(y,null,{title:e(()=>[t(n(a(s)("http.api.property.tags")),1)]),body:e(()=>[l(A,{tags:d.tags},null,8,["tags"])]),_:2},1024),t(),l(y,null,{title:e(()=>[t(n(a(s)("http.api.property.address")),1)]),body:e(()=>[l(K,{text:d.addressPort},null,8,["text"])]),_:2},1024),t(),l(y,null,{title:e(()=>[t(n(a(s)("http.api.property.serviceAddress")),1)]),body:e(()=>[l(K,{text:d.serviceAddressPort},null,8,["text"])]),_:2},1024)])]))),128))])]),_:1})])):$("",!0),t(),o("div",Pt,[o("h2",null,n(a(s)("data-planes.routes.item.mtls.title")),1),t(),u.data.dataplaneInsight.mTLS?(r(!0),c(m,{key:0},x([u.data.dataplaneInsight.mTLS],d=>(r(),f(I,{key:d,class:"mt-4"},{default:e(()=>[o("div",Rt,[l(y,null,{title:e(()=>[t(n(a(s)("data-planes.routes.item.mtls.expiration_time.title")),1)]),body:e(()=>[t(n(a(p)(d.certificateExpirationTime)),1)]),_:2},1024),t(),l(y,null,{title:e(()=>[t(n(a(s)("data-planes.routes.item.mtls.generation_time.title")),1)]),body:e(()=>[t(n(a(p)(d.lastCertificateRegeneration)),1)]),_:2},1024),t(),l(y,null,{title:e(()=>[t(n(a(s)("data-planes.routes.item.mtls.regenerations.title")),1)]),body:e(()=>[t(n(a(s)("common.formats.integer",{value:d.certificateRegenerations})),1)]),_:2},1024),t(),l(y,null,{title:e(()=>[t(n(a(s)("data-planes.routes.item.mtls.issued_backend.title")),1)]),body:e(()=>[t(n(d.issuedBackend),1)]),_:2},1024),t(),l(y,null,{title:e(()=>[t(n(a(s)("data-planes.routes.item.mtls.supported_backends.title")),1)]),body:e(()=>[o("ul",null,[(r(!0),c(m,null,x(d.supportedBackends,k=>(r(),c("li",{key:k},n(k),1))),128))])]),_:2},1024)])]),_:2},1024))),128)):(r(),f(H,{key:1,class:"mt-4",appearance:"warning"},{alertMessage:e(()=>[o("div",{innerHTML:a(s)("data-planes.routes.item.mtls.disabled")},null,8,Mt)]),_:1}))]),t(),u.data.dataplaneInsight.subscriptions.length>0?(r(),c("div",Vt,[o("h2",null,n(a(s)("data-planes.routes.item.subscriptions.title")),1),t(),l(I,{class:"mt-4"},{default:e(()=>[l(rt,{subscriptions:u.data.dataplaneInsight.subscriptions},null,8,["subscriptions"])]),_:1})])):$("",!0)])]),_:2},[C.value.length>0?{name:"notifications",fn:e(()=>[o("ul",xt,[(r(!0),c(m,null,x(C.value,d=>(r(),c("li",{key:d.kind,"data-testid":`warning-${d.kind}`,innerHTML:a(s)(`common.warnings.${d.kind}`,d.payload)},null,8,kt))),128)),t()])]),key:"0"}:void 0]),1024)]),_:1})}}});const Zt=D(Et,[["__scopeId","data-v-f8375696"]]);export{Zt as default};
