import{a as j,K as L}from"./index-fce48c05.js";import{_ as V,a as h,o as r,b as g,w as e,m as a,r as S,f as t,d as A,c as i,e as d,k as W,t as o,l as n,F as _,C as b,p as x,n as J,G as Y,V as G,W as k,X as Z,Y as Q,$ as tt,a0 as et,B as at,v as nt,x as st}from"./index-12461953.js";import{S as ot}from"./StatusBadge-57924c14.js";import{S as lt}from"./SummaryView-fa8fa616.js";import{T as H}from"./TagList-d671d1ac.js";import{T as rt}from"./TextWithCopyButton-9a13e075.js";import{_ as dt}from"./SubscriptionList.vue_vue_type_script_setup_true_lang-818642fb.js";import"./CopyButton-7dcbb455.js";import"./AccordionList-f1733260.js";const it={},ct={class:"card"},ut={class:"title"},_t={class:"body"};function pt(v,s){const u=h("KCard");return r(),g(u,{class:"data-card"},{default:e(()=>[a("dl",null,[a("div",ct,[a("dt",ut,[S(v.$slots,"title",{},void 0,!0)]),t(),a("dd",_t,[S(v.$slots,"default",{},void 0,!0)])])])]),_:3})}const X=V(it,[["render",pt],["__scopeId","data-v-6e083223"]]),mt={class:"service-traffic"},ft={class:"actions"},vt=A({__name:"DataPlaneTraffic",setup(v){return(s,u)=>(r(),i("div",mt,[a("div",ft,[S(s.$slots,"actions",{},void 0,!0)]),t(),d(X,{class:"header"},{title:e(()=>[S(s.$slots,"title",{},void 0,!0)]),_:3}),t(),S(s.$slots,"default",{},void 0,!0)]))}});const F=V(vt,[["__scopeId","data-v-5bd1dbf9"]]),yt={class:"title"},ht={key:0},gt=A({__name:"ServiceTrafficCard",props:{protocol:{},traffic:{}},setup(v){const{t:s}=W(),u=v,p=w=>{const B=w.target;if(B.nodeName.toLowerCase()!=="a"){const T=B.closest(".service-traffic-card");if(T){const $=T.querySelector("a");$!==null&&$.click()}}};return(w,B)=>{const T=h("KBadge");return r(),g(X,{class:"service-traffic-card",onClick:p},{title:e(()=>[d(T,{appearance:u.protocol==="passthrough"?"success":"info"},{default:e(()=>[t(o(n(s)(`data-planes.components.service_traffic_card.protocol.${u.protocol}`,{},{defaultMessage:n(s)(`http.api.value.${u.protocol}`)})),1)]),_:1},8,["appearance"]),t(),a("div",yt,[S(w.$slots,"default",{},void 0,!0)])]),default:e(()=>{var $,C,I,R,P,q,z,E,N;return[t(),u.traffic?(r(),i("dl",ht,[u.protocol==="passthrough"?(r(!0),i(_,{key:0},b([["http","tcp"].reduce((f,m)=>{var D;const y="downstream";return Object.entries(((D=u.traffic)==null?void 0:D[m])||{}).reduce((K,[l,c])=>[`${y}_cx_tx_bytes_total`,`${y}_cx_rx_bytes_total`].includes(l)?{...K,[l]:c+(K[l]??0)}:K,f)},{})],(f,m)=>(r(),i(_,{key:m},[a("div",null,[a("dt",null,o(n(s)("data-planes.components.service_traffic_card.tx")),1),t(),a("dd",null,o(n(s)("common.formats.bytes",{value:f.downstream_cx_rx_bytes_total??0})),1)]),t(),a("div",null,[a("dt",null,o(n(s)("data-planes.components.service_traffic_card.rx")),1),t(),a("dd",null,o(n(s)("common.formats.bytes",{value:f.downstream_cx_tx_bytes_total??0})),1)])],64))),128)):u.protocol==="grpc"?(r(),i(_,{key:1},[a("div",null,[a("dt",null,o(n(s)("data-planes.components.service_traffic_card.grpc_success")),1),t(),a("dd",null,o(n(s)("common.formats.integer",{value:(($=u.traffic.grpc)==null?void 0:$.success)??0})),1)]),t(),a("div",null,[a("dt",null,o(n(s)("data-planes.components.service_traffic_card.grpc_failure")),1),t(),a("dd",null,o(n(s)("common.formats.integer",{value:((C=u.traffic.grpc)==null?void 0:C.failure)??0})),1)])],64)):u.protocol==="http"?(r(),i(_,{key:2},[(r(!0),i(_,null,b([((I=u.traffic.http)==null?void 0:I.downstream_rq_1xx)??0].filter(f=>f!==0),f=>(r(),i("div",{key:f},[a("dt",null,o(n(s)("data-planes.components.service_traffic_card.1xx")),1),t(),a("dd",null,o(n(s)("common.formats.integer",{value:f})),1)]))),128)),t(),a("div",null,[a("dt",null,o(n(s)("data-planes.components.service_traffic_card.2xx")),1),t(),a("dd",null,o(n(s)("common.formats.integer",{value:((R=u.traffic.http)==null?void 0:R.downstream_rq_2xx)??0})),1)]),t(),(r(!0),i(_,null,b([((P=u.traffic.http)==null?void 0:P.downstream_rq_3xx)??0].filter(f=>f!==0),f=>(r(),i("div",{key:f},[a("dt",null,o(n(s)("data-planes.components.service_traffic_card.3xx")),1),t(),a("dd",null,o(n(s)("common.formats.integer",{value:f})),1)]))),128)),t(),a("div",null,[a("dt",null,o(n(s)("data-planes.components.service_traffic_card.4xx")),1),t(),a("dd",null,o(n(s)("common.formats.integer",{value:((q=u.traffic.http)==null?void 0:q.downstream_rq_4xx)??0})),1)]),t(),a("div",null,[a("dt",null,o(n(s)("data-planes.components.service_traffic_card.5xx")),1),t(),a("dd",null,o(n(s)("common.formats.integer",{value:((z=u.traffic.http)==null?void 0:z.downstream_rq_5xx)??0})),1)])],64)):(r(),i(_,{key:3},[a("div",null,[a("dt",null,o(n(s)("data-planes.components.service_traffic_card.tx")),1),t(),a("dd",null,o(n(s)("common.formats.bytes",{value:((E=u.traffic.tcp)==null?void 0:E.downstream_cx_rx_bytes_total)??0})),1)]),t(),a("div",null,[a("dt",null,o(n(s)("data-planes.components.service_traffic_card.rx")),1),t(),a("dd",null,o(n(s)("common.formats.bytes",{value:((N=u.traffic.tcp)==null?void 0:N.downstream_cx_tx_bytes_total)??0})),1)])],64))])):x("",!0)]}),_:3})}}});const M=V(gt,[["__scopeId","data-v-eddcb161"]]),bt={class:"body"},kt=A({__name:"ServiceTrafficGroup",props:{type:{}},setup(v){const s=v;return(u,p)=>{const w=h("KCard");return r(),g(w,{class:J(["service-traffic-group",`type-${s.type}`])},{default:e(()=>[a("div",bt,[S(u.$slots,"default",{},void 0,!0)])]),_:3},8,["class"])}}});const O=V(kt,[["__scopeId","data-v-baf4abf7"]]),xt=v=>(nt("data-v-f979a7ee"),v=v(),st(),v),wt={"data-testid":"dataplane-warnings"},$t=["data-testid","innerHTML"],Ct={key:0,"data-testid":"warning-stats-loading"},St={class:"stack","data-testid":"dataplane-details"},Tt={class:"columns"},It={class:"status-with-reason"},Dt={class:"columns"},Kt=xt(()=>a("span",null,"Outbounds",-1)),Vt={"data-testid":"dataplane-mtls"},Bt={class:"columns"},Rt=["innerHTML"],Pt={key:2,"data-testid":"dataplane-subscriptions"},qt=A({__name:"DataPlaneDetailView",props:{data:{}},setup(v){const{t:s,formatIsoDate:u}=W(),p=v,w=Y(()=>p.data.warnings.concat(...p.data.isCertExpired?[{kind:"CERT_EXPIRED"}]:[]));return(B,T)=>{const $=h("KTooltip"),C=h("KCard"),I=h("RouterLink"),R=h("KInputSwitch"),P=h("KButton"),q=h("RouterView"),z=h("KAlert"),E=h("AppView"),N=h("DataSource"),f=h("RouteView");return r(),g(f,{params:{mesh:"",dataPlane:"",service:"",inactive:!1},name:"data-plane-detail-view"},{default:e(({route:m})=>[d(N,{src:p.data.dataplaneType==="standard"?`/meshes/${m.params.mesh}/dataplanes/${m.params.dataPlane}/traffic`:""},{default:e(({data:y,error:D,refresh:K})=>[d(E,null,G({default:e(()=>[t(),a("div",St,[d(C,null,{default:e(()=>[a("div",Tt,[d(k,null,{title:e(()=>[t(o(n(s)("http.api.property.status")),1)]),body:e(()=>[a("div",It,[d(ot,{status:p.data.status},null,8,["status"]),t(),(r(!0),i(_,null,b([p.data.dataplane.networking.inbounds.filter(l=>!l.health.ready)],l=>(r(),i(_,{key:l},[l.length>0?(r(),g($,{key:0,class:"reason-tooltip","position-fixed":""},{content:e(()=>[a("ul",null,[(r(!0),i(_,null,b(l,c=>(r(),i("li",{key:`${c.service}:${c.port}`},o(n(s)("data-planes.routes.item.unhealthy_inbound",{service:c.service,port:c.port})),1))),128))])]),default:e(()=>[d(n(Z),{color:n(j),size:n(L),"hide-title":""},null,8,["color","size"]),t()]),_:2},1024)):x("",!0)],64))),128))])]),_:1}),t(),d(k,null,{title:e(()=>[t(o(n(s)("data-planes.routes.item.last_updated")),1)]),body:e(()=>[t(o(n(u)(p.data.modificationTime)),1)]),_:1}),t(),p.data.dataplane.networking.gateway?(r(),i(_,{key:0},[d(k,null,{title:e(()=>[t(o(n(s)("http.api.property.tags")),1)]),body:e(()=>[d(H,{tags:p.data.dataplane.networking.gateway.tags},null,8,["tags"])]),_:1}),t(),d(k,null,{title:e(()=>[t(o(n(s)("http.api.property.address")),1)]),body:e(()=>[d(rt,{text:`${p.data.dataplane.networking.address}`},null,8,["text"])]),_:1})],64)):x("",!0)])]),_:1}),t(),p.data.dataplaneType==="standard"?(r(),g(C,{key:0,class:"traffic","data-testid":"dataplane-traffic"},{default:e(()=>[a("div",Dt,[d(F,null,{title:e(()=>[d(n(Q),{display:"inline-block",decorative:"",size:n(L)},null,8,["size"]),t(`
                  Inbounds
                `)]),default:e(()=>[t(),d(O,{type:"inbound"},{default:e(()=>[(r(!0),i(_,null,b(p.data.dataplane.networking.inbounds,l=>(r(),i(_,{key:`${l.port}`},[(r(!0),i(_,null,b([(y||{inbounds:[]}).inbounds.find(c=>`${c.port}`==`${l.port}`)],c=>(r(),g(M,{key:c,protocol:l.protocol,traffic:c},{default:e(()=>[d(I,{to:{name:"data-plane-inbound-summary-overview-view",params:{service:l.port},query:{inactive:m.params.inactive?null:void 0}}},{default:e(()=>[t(`
                          :`+o(l.port),1)]),_:2},1032,["to"]),t(),d(H,{tags:[{label:"kuma.io/service",value:l.tags["kuma.io/service"]}]},null,8,["tags"])]),_:2},1032,["protocol","traffic"]))),128))],64))),128))]),_:2},1024)]),_:2},1024),t(),d(F,null,G({title:e(()=>[d(n(tt),{display:"inline-block",decorative:"",size:n(L)},null,8,["size"]),t(),Kt]),default:e(()=>[t(),t(),y?(r(),i(_,{key:0},[d(O,{type:"passthrough"},{default:e(()=>[d(M,{protocol:"passthrough",traffic:y.passthrough},{default:e(()=>[t(`
                      Non mesh traffic
                    `)]),_:2},1032,["traffic"])]),_:2},1024),t(),(r(!0),i(_,null,b([m.params.inactive?y.outbounds:y.outbounds.filter(l=>{var c,U;return(l.protocol==="tcp"?(c=l.tcp)==null?void 0:c.downstream_cx_rx_bytes_total:(U=l.http)==null?void 0:U.downstream_rq_total)??0>0})],l=>(r(),i(_,{key:l},[l.length>0?(r(),g(O,{key:0,type:"outbound","data-testid":"dataplane-outbounds"},{default:e(()=>[(r(!0),i(_,null,b(l,c=>(r(),g(M,{key:`${c.name}`,protocol:c.protocol,traffic:c},{default:e(()=>[d(I,{to:{name:"data-plane-outbound-summary-overview-view",params:{service:c.name},query:{inactive:m.params.inactive?null:void 0}}},{default:e(()=>[t(o(c.name),1)]),_:2},1032,["to"])]),_:2},1032,["protocol","traffic"]))),128))]),_:2},1024)):x("",!0)],64))),128))],64)):x("",!0)]),_:2},[y?{name:"actions",fn:e(()=>[d(R,{modelValue:m.params.inactive,"onUpdate:modelValue":l=>m.params.inactive=l,"data-testid":"dataplane-outbounds-inactive-toggle"},{label:e(()=>[t(`
                      Show inactive
                    `)]),_:2},1032,["modelValue","onUpdate:modelValue"]),t(),d(P,{appearance:"primary",onClick:K},{default:e(()=>[d(n(et),{size:n(L)},null,8,["size"]),t(`

                    Refresh
                  `)]),_:2},1032,["onClick"])]),key:"0"}:void 0]),1024)])]),_:2},1024)):x("",!0),t(),m.params.service&&[y==null?void 0:y.outbounds,p.data.dataplane.networking.inbounds].every(l=>typeof l<"u")?(r(),g(q,{key:1},{default:e(l=>[d(lt,{onClose:function(c){m.replace({name:"data-plane-detail-view",params:{mesh:m.params.mesh,dataPlane:m.params.dataPlane},query:{inactive:m.params.inactive?null:void 0}})}},{default:e(()=>[(r(),g(at(l.Component),{data:String(l.route.name).includes("-inbound-")?p.data.dataplane.networking.inbounds.find(c=>`${c.port}`===m.params.service):y.outbounds.find(c=>c.name===m.params.service)},null,8,["data"]))]),_:2},1032,["onClose"])]),_:2},1024)):x("",!0),t(),a("div",Vt,[a("h2",null,o(n(s)("data-planes.routes.item.mtls.title")),1),t(),p.data.dataplaneInsight.mTLS?(r(!0),i(_,{key:0},b([p.data.dataplaneInsight.mTLS],l=>(r(),g(C,{key:l,class:"mt-4"},{default:e(()=>[a("div",Bt,[d(k,null,{title:e(()=>[t(o(n(s)("data-planes.routes.item.mtls.expiration_time.title")),1)]),body:e(()=>[t(o(n(u)(l.certificateExpirationTime)),1)]),_:2},1024),t(),d(k,null,{title:e(()=>[t(o(n(s)("data-planes.routes.item.mtls.generation_time.title")),1)]),body:e(()=>[t(o(n(u)(l.lastCertificateRegeneration)),1)]),_:2},1024),t(),d(k,null,{title:e(()=>[t(o(n(s)("data-planes.routes.item.mtls.regenerations.title")),1)]),body:e(()=>[t(o(n(s)("common.formats.integer",{value:l.certificateRegenerations})),1)]),_:2},1024),t(),d(k,null,{title:e(()=>[t(o(n(s)("data-planes.routes.item.mtls.issued_backend.title")),1)]),body:e(()=>[t(o(l.issuedBackend),1)]),_:2},1024),t(),d(k,null,{title:e(()=>[t(o(n(s)("data-planes.routes.item.mtls.supported_backends.title")),1)]),body:e(()=>[a("ul",null,[(r(!0),i(_,null,b(l.supportedBackends,c=>(r(),i("li",{key:c},o(c),1))),128))])]),_:2},1024)])]),_:2},1024))),128)):(r(),g(z,{key:1,class:"mt-4",appearance:"warning"},{alertMessage:e(()=>[a("div",{innerHTML:n(s)("data-planes.routes.item.mtls.disabled")},null,8,Rt)]),_:1}))]),t(),p.data.dataplaneInsight.subscriptions.length>0?(r(),i("div",Pt,[a("h2",null,o(n(s)("data-planes.routes.item.subscriptions.title")),1),t(),d(C,{class:"mt-4"},{default:e(()=>[d(dt,{subscriptions:p.data.dataplaneInsight.subscriptions},null,8,["subscriptions"])]),_:1})])):x("",!0)])]),_:2},[w.value.length>0?{name:"notifications",fn:e(()=>[a("ul",wt,[(r(!0),i(_,null,b(w.value,l=>(r(),i("li",{key:l.kind,"data-testid":`warning-${l.kind}`,innerHTML:n(s)(`common.warnings.${l.kind}`,l.payload)},null,8,$t))),128)),t(),D?(r(),i("li",Ct,[t(`
              Error loading outbound stats: `),a("strong",null,o(D.toString()),1)])):x("",!0),t()])]),key:"0"}:void 0]),1024)]),_:2},1032,["src"])]),_:1})}}});const Ht=V(qt,[["__scopeId","data-v-f979a7ee"]]);export{Ht as default};
