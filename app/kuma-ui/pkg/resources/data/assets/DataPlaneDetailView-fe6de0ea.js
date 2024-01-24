import{a as j,K as A}from"./index-fce48c05.js";import{_ as P,a as h,o as l,b as g,w as a,m as n,r as S,f as t,d as O,c as i,e as d,k as X,t as o,l as e,F as p,C as b,p as w,n as Y,y as J,H as Q,W as G,X as k,Y as tt,Z as et,$ as at,a0 as nt,B as st,s as ot,v as lt}from"./index-f4426e51.js";import{S as rt}from"./StatusBadge-0b40c424.js";import{S as dt}from"./SummaryView-d3e8ae1d.js";import{T as W}from"./TagList-bb744c26.js";import{T as it}from"./TextWithCopyButton-a0c3a6c2.js";import{_ as ct}from"./SubscriptionList.vue_vue_type_script_setup_true_lang-5504f75d.js";import"./CopyButton-2e20cd8b.js";import"./AccordionList-54af0d27.js";const ut={},_t={class:"card"},pt={class:"title"},mt={class:"body"};function ft(v,s){const c=h("KCard");return l(),g(c,{class:"data-card"},{default:a(()=>[n("dl",null,[n("div",_t,[n("dt",pt,[S(v.$slots,"title",{},void 0,!0)]),t(),n("dd",mt,[S(v.$slots,"default",{},void 0,!0)])])])]),_:3})}const Z=P(ut,[["render",ft],["__scopeId","data-v-6e083223"]]),vt={class:"service-traffic"},yt={class:"actions"},ht=O({__name:"DataPlaneTraffic",setup(v){return(s,c)=>(l(),i("div",vt,[n("div",yt,[S(s.$slots,"actions",{},void 0,!0)]),t(),d(Z,{class:"header"},{title:a(()=>[S(s.$slots,"title",{},void 0,!0)]),_:3}),t(),S(s.$slots,"default",{},void 0,!0)]))}});const F=P(ht,[["__scopeId","data-v-5bd1dbf9"]]),gt={class:"title"},bt={key:0},kt=O({__name:"ServiceTrafficCard",props:{protocol:{},traffic:{}},setup(v){const{t:s}=X(),c=v,K=_=>{const T=_.target;if(T.nodeName.toLowerCase()!=="a"){const D=T.closest(".service-traffic-card");if(D){const I=D.querySelector("a");I!==null&&I.click()}}};return(_,T)=>{const D=h("KBadge");return l(),g(Z,{class:"service-traffic-card",onClick:K},{title:a(()=>[d(D,{appearance:c.protocol==="passthrough"?"success":"info"},{default:a(()=>[t(o(e(s)(`data-planes.components.service_traffic_card.protocol.${c.protocol}`,{},{defaultMessage:e(s)(`http.api.value.${c.protocol}`)})),1)]),_:1},8,["appearance"]),t(),n("div",gt,[S(_.$slots,"default",{},void 0,!0)])]),default:a(()=>{var I,q,$,V,z,E,N,L,M;return[t(),c.traffic?(l(),i("dl",bt,[c.protocol==="passthrough"?(l(!0),i(p,{key:0},b([["http","tcp"].reduce((f,R)=>{var y;const m="downstream";return Object.entries(((y=c.traffic)==null?void 0:y[R])||{}).reduce((C,[B,r])=>[`${m}_cx_tx_bytes_total`,`${m}_cx_rx_bytes_total`].includes(B)?{...C,[B]:r+(C[B]??0)}:C,f)},{})],(f,R)=>(l(),i(p,{key:R},[n("div",null,[n("dt",null,o(e(s)("data-planes.components.service_traffic_card.tx")),1),t(),n("dd",null,o(e(s)("common.formats.bytes",{value:f.downstream_cx_rx_bytes_total??0})),1)]),t(),n("div",null,[n("dt",null,o(e(s)("data-planes.components.service_traffic_card.rx")),1),t(),n("dd",null,o(e(s)("common.formats.bytes",{value:f.downstream_cx_tx_bytes_total??0})),1)])],64))),128)):c.protocol==="grpc"?(l(),i(p,{key:1},[n("div",null,[n("dt",null,o(e(s)("data-planes.components.service_traffic_card.grpc_success")),1),t(),n("dd",null,o(e(s)("common.formats.integer",{value:((I=c.traffic.grpc)==null?void 0:I.success)??0})),1)]),t(),n("div",null,[n("dt",null,o(e(s)("data-planes.components.service_traffic_card.grpc_failure")),1),t(),n("dd",null,o(e(s)("common.formats.integer",{value:((q=c.traffic.grpc)==null?void 0:q.failure)??0})),1)])],64)):c.protocol==="http"?(l(),i(p,{key:2},[(l(!0),i(p,null,b([(($=c.traffic.http)==null?void 0:$.downstream_rq_1xx)??0].filter(f=>f!==0),f=>(l(),i("div",{key:f},[n("dt",null,o(e(s)("data-planes.components.service_traffic_card.1xx")),1),t(),n("dd",null,o(e(s)("common.formats.integer",{value:f})),1)]))),128)),t(),n("div",null,[n("dt",null,o(e(s)("data-planes.components.service_traffic_card.2xx")),1),t(),n("dd",null,o(e(s)("common.formats.integer",{value:((V=c.traffic.http)==null?void 0:V.downstream_rq_2xx)??0})),1)]),t(),(l(!0),i(p,null,b([((z=c.traffic.http)==null?void 0:z.downstream_rq_3xx)??0].filter(f=>f!==0),f=>(l(),i("div",{key:f},[n("dt",null,o(e(s)("data-planes.components.service_traffic_card.3xx")),1),t(),n("dd",null,o(e(s)("common.formats.integer",{value:f})),1)]))),128)),t(),n("div",null,[n("dt",null,o(e(s)("data-planes.components.service_traffic_card.4xx")),1),t(),n("dd",null,o(e(s)("common.formats.integer",{value:((E=c.traffic.http)==null?void 0:E.downstream_rq_4xx)??0})),1)]),t(),n("div",null,[n("dt",null,o(e(s)("data-planes.components.service_traffic_card.5xx")),1),t(),n("dd",null,o(e(s)("common.formats.integer",{value:((N=c.traffic.http)==null?void 0:N.downstream_rq_5xx)??0})),1)])],64)):(l(),i(p,{key:3},[n("div",null,[n("dt",null,o(e(s)("data-planes.components.service_traffic_card.tx")),1),t(),n("dd",null,o(e(s)("common.formats.bytes",{value:((L=c.traffic.tcp)==null?void 0:L.downstream_cx_rx_bytes_total)??0})),1)]),t(),n("div",null,[n("dt",null,o(e(s)("data-planes.components.service_traffic_card.rx")),1),t(),n("dd",null,o(e(s)("common.formats.bytes",{value:((M=c.traffic.tcp)==null?void 0:M.downstream_cx_tx_bytes_total)??0})),1)])],64))])):w("",!0)]}),_:3})}}});const U=P(kt,[["__scopeId","data-v-eddcb161"]]),wt={class:"body"},xt=O({__name:"ServiceTrafficGroup",props:{type:{}},setup(v){const s=v;return(c,K)=>{const _=h("KCard");return l(),g(_,{class:Y(["service-traffic-group",`type-${s.type}`])},{default:a(()=>[n("div",wt,[S(c.$slots,"default",{},void 0,!0)])]),_:3},8,["class"])}}});const H=P(xt,[["__scopeId","data-v-baf4abf7"]]),$t=v=>(ot("data-v-4683403d"),v=v(),lt(),v),Ct={"data-testid":"dataplane-warnings"},St=["data-testid","innerHTML"],Tt={key:0,"data-testid":"warning-stats-loading"},It={class:"stack","data-testid":"dataplane-details"},Kt={class:"columns"},Dt={class:"status-with-reason"},Vt={class:"columns"},Rt=$t(()=>n("span",null,"Outbounds",-1)),Bt={"data-testid":"dataplane-mtls"},Pt={class:"columns"},qt=["innerHTML"],zt={key:2,"data-testid":"dataplane-subscriptions"},Et=O({__name:"DataPlaneDetailView",props:{data:{}},setup(v){const{t:s,formatIsoDate:c}=X(),K=J(),_=v,T=Q(()=>_.data.warnings.concat(..._.data.isCertExpired?[{kind:"CERT_EXPIRED"}]:[]));return(D,I)=>{const q=h("KTooltip"),$=h("KCard"),V=h("RouterLink"),z=h("KInputSwitch"),E=h("KButton"),N=h("RouterView"),L=h("KAlert"),M=h("AppView"),f=h("DataSource"),R=h("RouteView");return l(),g(R,{params:{mesh:"",dataPlane:"",service:"",inactive:!1},name:"data-plane-detail-view"},{default:a(({route:m})=>[d(f,{src:_.data.dataplaneType==="standard"?`/meshes/${m.params.mesh}/dataplanes/${m.params.dataPlane}/stats`:""},{default:a(({data:y,error:C,refresh:B})=>[d(M,null,G({default:a(()=>[t(),n("div",It,[d($,null,{default:a(()=>[n("div",Kt,[d(k,null,{title:a(()=>[t(o(e(s)("http.api.property.status")),1)]),body:a(()=>[n("div",Dt,[d(rt,{status:_.data.status},null,8,["status"]),t(),(l(!0),i(p,null,b([_.data.dataplane.networking.inbounds.filter(r=>!r.health.ready)],r=>(l(),i(p,{key:r},[r.length>0?(l(),g(q,{key:0,class:"reason-tooltip","position-fixed":""},{content:a(()=>[n("ul",null,[(l(!0),i(p,null,b(r,u=>(l(),i("li",{key:`${u.service}:${u.port}`},o(e(s)("data-planes.routes.item.unhealthy_inbound",{service:u.service,port:u.port})),1))),128))])]),default:a(()=>[d(e(tt),{color:e(j),size:e(A),"hide-title":""},null,8,["color","size"]),t()]),_:2},1024)):w("",!0)],64))),128))])]),_:1}),t(),d(k,null,{title:a(()=>[t(o(e(s)("data-planes.routes.item.last_updated")),1)]),body:a(()=>[t(o(e(c)(_.data.modificationTime)),1)]),_:1}),t(),_.data.dataplane.networking.gateway?(l(),i(p,{key:0},[d(k,null,{title:a(()=>[t(o(e(s)("http.api.property.tags")),1)]),body:a(()=>[d(W,{tags:_.data.dataplane.networking.gateway.tags},null,8,["tags"])]),_:1}),t(),d(k,null,{title:a(()=>[t(o(e(s)("http.api.property.address")),1)]),body:a(()=>[d(it,{text:`${_.data.dataplane.networking.address}`},null,8,["text"])]),_:1})],64)):w("",!0)])]),_:1}),t(),_.data.dataplaneType==="standard"?(l(),g($,{key:0,class:"traffic","data-testid":"dataplane-traffic"},{default:a(()=>[n("div",Vt,[d(F,null,{title:a(()=>[d(e(et),{display:"inline-block",decorative:"",size:e(A)},null,8,["size"]),t(`
                  Inbounds
                `)]),default:a(()=>[t(),d(H,{type:"inbound"},{default:a(()=>[(l(!0),i(p,null,b(_.data.dataplane.networking.inbounds,r=>(l(),i(p,{key:`${r.port}`},[(l(!0),i(p,null,b([(y||{inbounds:[]}).inbounds.find(u=>`${u.port}`==`${r.port}`)],u=>(l(),g(U,{key:u,protocol:r.protocol,traffic:u},{default:a(()=>[d(V,{to:{name:(x=>x.includes("bound")?x.replace("-outbound-","-inbound-"):"data-plane-inbound-summary-overview-view")(String(e(K).name)),params:{service:r.port},query:{inactive:m.params.inactive?null:void 0}}},{default:a(()=>[t(`
                          :`+o(r.port),1)]),_:2},1032,["to"]),t(),d(W,{tags:[{label:"kuma.io/service",value:r.tags["kuma.io/service"]}]},null,8,["tags"])]),_:2},1032,["protocol","traffic"]))),128))],64))),128))]),_:2},1024)]),_:2},1024),t(),d(F,null,G({title:a(()=>[d(e(at),{display:"inline-block",decorative:"",size:e(A)},null,8,["size"]),t(),Rt]),default:a(()=>[t(),t(),y?(l(),i(p,{key:0},[d(H,{type:"passthrough"},{default:a(()=>[d(U,{protocol:"passthrough",traffic:y.passthrough},{default:a(()=>[t(`
                      Non mesh traffic
                    `)]),_:2},1032,["traffic"])]),_:2},1024),t(),(l(!0),i(p,null,b([m.params.inactive?y.outbounds:y.outbounds.filter(r=>{var u,x;return(r.protocol==="tcp"?(u=r.tcp)==null?void 0:u.downstream_cx_rx_bytes_total:(x=r.http)==null?void 0:x.downstream_rq_total)??0>0})],r=>(l(),i(p,{key:r},[r.length>0?(l(),g(H,{key:0,type:"outbound","data-testid":"dataplane-outbounds"},{default:a(()=>[(l(!0),i(p,null,b(r,u=>(l(),g(U,{key:`${u.name}`,protocol:u.protocol,traffic:u},{default:a(()=>[d(V,{to:{name:(x=>x.includes("bound")?x.replace("-inbound-","-outbound-"):"data-plane-outbound-summary-overview-view")(String(e(K).name)),params:{service:u.name},query:{inactive:m.params.inactive?null:void 0}}},{default:a(()=>[t(o(u.name),1)]),_:2},1032,["to"])]),_:2},1032,["protocol","traffic"]))),128))]),_:2},1024)):w("",!0)],64))),128))],64)):w("",!0)]),_:2},[y?{name:"actions",fn:a(()=>[d(z,{modelValue:m.params.inactive,"onUpdate:modelValue":r=>m.params.inactive=r,"data-testid":"dataplane-outbounds-inactive-toggle"},{label:a(()=>[t(`
                      Show inactive
                    `)]),_:2},1032,["modelValue","onUpdate:modelValue"]),t(),d(E,{appearance:"primary",onClick:B},{default:a(()=>[d(e(nt),{size:e(A)},null,8,["size"]),t(`

                    Refresh
                  `)]),_:2},1032,["onClick"])]),key:"0"}:void 0]),1024)])]),_:2},1024)):w("",!0),t(),m.params.service&&[y==null?void 0:y.outbounds,_.data.dataplane.networking.inbounds].every(r=>typeof r<"u")?(l(),g(N,{key:1},{default:a(r=>[d(dt,{width:"670px",onClose:function(u){m.replace({name:"data-plane-detail-view",params:{mesh:m.params.mesh,dataPlane:m.params.dataPlane},query:{inactive:m.params.inactive?null:void 0}})}},{default:a(()=>[(l(),g(st(r.Component),{data:String(r.route.name).includes("-inbound-")?_.data.dataplane.networking.inbounds.find(u=>`${u.port}`===m.params.service):y.outbounds.find(u=>u.name===m.params.service)},null,8,["data"]))]),_:2},1032,["onClose"])]),_:2},1024)):w("",!0),t(),n("div",Bt,[n("h2",null,o(e(s)("data-planes.routes.item.mtls.title")),1),t(),_.data.dataplaneInsight.mTLS?(l(!0),i(p,{key:0},b([_.data.dataplaneInsight.mTLS],r=>(l(),g($,{key:r,class:"mt-4"},{default:a(()=>[n("div",Pt,[d(k,null,{title:a(()=>[t(o(e(s)("data-planes.routes.item.mtls.expiration_time.title")),1)]),body:a(()=>[t(o(e(c)(r.certificateExpirationTime)),1)]),_:2},1024),t(),d(k,null,{title:a(()=>[t(o(e(s)("data-planes.routes.item.mtls.generation_time.title")),1)]),body:a(()=>[t(o(e(c)(r.lastCertificateRegeneration)),1)]),_:2},1024),t(),d(k,null,{title:a(()=>[t(o(e(s)("data-planes.routes.item.mtls.regenerations.title")),1)]),body:a(()=>[t(o(e(s)("common.formats.integer",{value:r.certificateRegenerations})),1)]),_:2},1024),t(),d(k,null,{title:a(()=>[t(o(e(s)("data-planes.routes.item.mtls.issued_backend.title")),1)]),body:a(()=>[t(o(r.issuedBackend),1)]),_:2},1024),t(),d(k,null,{title:a(()=>[t(o(e(s)("data-planes.routes.item.mtls.supported_backends.title")),1)]),body:a(()=>[n("ul",null,[(l(!0),i(p,null,b(r.supportedBackends,u=>(l(),i("li",{key:u},o(u),1))),128))])]),_:2},1024)])]),_:2},1024))),128)):(l(),g(L,{key:1,class:"mt-4",appearance:"warning"},{alertMessage:a(()=>[n("div",{innerHTML:e(s)("data-planes.routes.item.mtls.disabled")},null,8,qt)]),_:1}))]),t(),_.data.dataplaneInsight.subscriptions.length>0?(l(),i("div",zt,[n("h2",null,o(e(s)("data-planes.routes.item.subscriptions.title")),1),t(),d($,{class:"mt-4"},{default:a(()=>[d(ct,{subscriptions:_.data.dataplaneInsight.subscriptions},null,8,["subscriptions"])]),_:1})])):w("",!0)])]),_:2},[T.value.length>0||C?{name:"notifications",fn:a(()=>[n("ul",Ct,[(l(!0),i(p,null,b(T.value,r=>(l(),i("li",{key:r.kind,"data-testid":`warning-${r.kind}`,innerHTML:e(s)(`common.warnings.${r.kind}`,r.payload)},null,8,St))),128)),t(),C?(l(),i("li",Tt,[t(`
              The below view is not enhanced with runtime stats (Error loading stats: `),n("strong",null,o(C.toString()),1),t(`)
            `)])):w("",!0),t()])]),key:"0"}:void 0]),1024)]),_:2},1032,["src"])]),_:1})}}});const Ft=P(Et,[["__scopeId","data-v-4683403d"]]);export{Ft as default};
