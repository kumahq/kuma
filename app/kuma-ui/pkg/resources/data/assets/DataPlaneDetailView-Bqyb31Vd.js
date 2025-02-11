import{d as tt,H as et,K as at,r as m,o as s,q as g,w as t,b as o,p as S,$ as nt,T as A,e as a,U as v,t as d,S as ot,m as w,c as u,M as y,N as _,s as C,I as st,B as dt,_ as it}from"./index-BP47cGGe.js";import{S as rt}from"./SummaryView-Cb9BKz7Y.js";import{T as lt}from"./TagList-B7fSNTxL.js";import{C as j,a as B,b as R}from"./ConnectionTraffic-F34rr1g0.js";const pt={"data-testid":"dataplane-warnings"},ut=["data-testid"],mt={key:0,"data-testid":"warning-stats-loading"},yt={"data-testid":"dataplane-mtls"},ct={class:"columns"},gt={key:0,"data-testid":"dataplane-subscriptions"},ft=tt({__name:"DataPlaneDetailView",props:{data:{},mesh:{}},setup(M){const V=et(),i=M,E=at(()=>i.data.warnings.concat(...i.data.isCertExpired?[{kind:"CERT_EXPIRED"}]:[]));return(bt,e)=>{const N=m("XI18n"),h=m("XIcon"),D=m("DataCollection"),X=m("XLayout"),z=m("XAction"),$=m("XBadge"),F=m("XCopyButton"),G=m("XAboutCard"),P=m("XEmptyState"),H=m("XInputSwitch"),K=m("XProgress"),q=m("XCard"),L=m("RouterView"),U=m("XAlert"),J=m("AppView"),Q=m("DataSource"),W=m("RouteView");return s(),g(W,{params:{inactive:Boolean,mesh:"",proxy:"",proxyType:"",subscription:""},name:"data-plane-detail-view"},{default:t(({route:c,t:r,can:O,me:x,uri:Y})=>[o(Q,{src:Y(S(nt),"/connections/stats/for/:proxyType/:name/:mesh/:socketAddress",{proxyType:{ingresses:"zone-ingress",egresses:"zone-egress"}[c.params.proxyType]??"dataplane",name:c.params.proxy,mesh:c.params.mesh||"*",socketAddress:i.data.dataplane.networking.inboundAddress})},{default:t(({data:b,error:T,refresh:Z})=>[o(J,null,A({default:t(()=>[e[48]||(e[48]=a()),o(X,{type:"stack","data-testid":"dataplane-details"},{default:t(()=>[o(G,{title:r("data-planes.routes.item.about.title"),created:i.data.creationTime,modified:i.data.modificationTime},{default:t(()=>[o(v,{layout:"horizontal"},{title:t(()=>[a(d(r("http.api.property.status")),1)]),body:t(()=>[o(X,{type:"separated"},{default:t(()=>[o(ot,{status:i.data.status},null,8,["status"]),e[3]||(e[3]=a()),i.data.dataplaneType==="standard"?(s(),g(D,{key:0,items:i.data.dataplane.networking.inbounds,predicate:n=>n.state!=="Ready",empty:!1},{default:t(({items:n})=>[o(h,{name:"info"},{default:t(()=>[w("ul",null,[(s(!0),u(y,null,_(n,p=>(s(),u("li",{key:`${p.service}:${p.port}`},d(r("data-planes.routes.item.unhealthy_inbound",{service:p.service,port:p.port})),1))),128))])]),_:2},1024)]),_:2},1032,["items","predicate"])):C("",!0)]),_:2},1024)]),_:2},1024),e[10]||(e[10]=a()),O("use zones")&&i.data.zone?(s(),g(v,{key:0,layout:"horizontal"},{title:t(()=>[a(d(r("http.api.property.zone")),1)]),body:t(()=>[o($,{appearance:"decorative"},{default:t(()=>[o(z,{to:{name:"zone-cp-detail-view",params:{zone:i.data.zone}}},{default:t(()=>[a(d(i.data.zone),1)]),_:1},8,["to"])]),_:1})]),_:2},1024)):C("",!0),e[11]||(e[11]=a()),o(v,{layout:"horizontal"},{title:t(()=>[a(d(r("http.api.proptery.type")),1)]),body:t(()=>[o($,{appearance:"decorative"},{default:t(()=>[a(d(r(`data-planes.type.${i.data.dataplaneType}`)),1)]),_:2},1024)]),_:2},1024),e[12]||(e[12]=a()),i.data.namespace.length>0?(s(),g(v,{key:1,layout:"horizontal"},{title:t(()=>[a(d(r("http.api.property.namespace")),1)]),body:t(()=>[o($,{appearance:"decorative"},{default:t(()=>[a(d(i.data.namespace),1)]),_:1})]),_:2},1024)):C("",!0),e[13]||(e[13]=a()),o(v,{layout:"horizontal"},{title:t(()=>[a(d(r("http.api.property.address")),1)]),body:t(()=>[o(F,{variant:"badge",format:"default",text:`${i.data.dataplane.networking.address}`},null,8,["text"])]),_:2},1024),e[14]||(e[14]=a()),i.data.dataplane.networking.gateway?(s(),g(v,{key:2,layout:"horizontal"},{title:t(()=>[a(d(r("http.api.property.tags")),1)]),body:t(()=>[o(lt,{tags:i.data.dataplane.networking.gateway.tags},null,8,["tags"])]),_:2},1024)):C("",!0)]),_:2},1032,["title","created","modified"]),e[44]||(e[44]=a()),o(q,{class:"traffic","data-testid":"dataplane-traffic"},{default:t(()=>[o(X,{type:"columns"},{default:t(()=>[o(j,null,{title:t(()=>[o(X,{type:"separated"},{default:t(()=>[o(h,{name:"inbound"}),e[15]||(e[15]=a()),e[16]||(e[16]=w("span",null,"Inbounds",-1))]),_:1})]),default:t(()=>[e[18]||(e[18]=a()),(s(!0),u(y,null,_([i.data.dataplane.networking.type==="gateway"?Object.entries((b==null?void 0:b.inbounds)??{}).reduce((n,[p,l])=>{var k;const f=p.split("_").at(-1);return f===(((k=i.data.dataplane.networking.admin)==null?void 0:k.port)??"9901")?n:n.concat([{...i.data.dataplane.networking.inbounds[0],name:p,port:Number(f),protocol:["http","tcp"].find(I=>typeof l[I]<"u")??"tcp",addressPort:`${i.data.dataplane.networking.inbounds[0].address}:${f}`}])},[]):i.data.dataplane.networking.inbounds],n=>(s(),g(B,{key:n,type:"inbound","data-testid":"dataplane-inbounds"},{default:t(()=>[o(D,{type:"inbounds",items:n,predicate:p=>p.port!==49151},A({default:t(({items:p})=>[o(X,{type:"stack",size:"small"},{default:t(()=>[(s(!0),u(y,null,_(p,l=>(s(),u(y,{key:`${l.name}`},[(s(!0),u(y,null,_([b==null?void 0:b.inbounds[l.name]],f=>(s(),g(R,{key:f,"data-testid":"dataplane-inbound",protocol:l.protocol,service:O("use service-insights",i.mesh)?l.tags["kuma.io/service"]:"","port-name":l.portName,traffic:typeof T>"u"?f:{name:"",protocol:l.protocol,port:`${l.port}`}},{default:t(()=>[o(z,{"data-action":"",to:{name:(k=>k.includes("bound")?k.replace("-outbound-","-inbound-"):"data-plane-connection-inbound-summary-overview-view")(String(S(V).name)),params:{connection:l.name},query:{inactive:c.params.inactive}}},{default:t(()=>[a(d(l.name.replace("localhost","").replace("_",":")),1)]),_:2},1032,["to"])]),_:2},1032,["protocol","service","port-name","traffic"]))),128))],64))),128))]),_:2},1024)]),_:2},[i.data.dataplaneType==="delegated"?{name:"empty",fn:t(()=>[o(P,null,{default:t(()=>[w("p",null,`
                            This proxy is a delegated gateway therefore `+d(r("common.product.name"))+` does not have any visibility into inbounds for this gateway.
                          `,1)]),_:2},1024)]),key:"0"}:void 0]),1032,["items","predicate"])]),_:2},1024))),128))]),_:2},1024),e[28]||(e[28]=a()),o(j,null,A({title:t(()=>[o(X,{type:"separated"},{default:t(()=>[o(h,{name:"outbound"}),e[22]||(e[22]=a()),e[23]||(e[23]=w("span",null,"Outbounds",-1))]),_:1})]),default:t(()=>[e[26]||(e[26]=a()),e[27]||(e[27]=a()),typeof T>"u"?(s(),u(y,{key:0},[typeof b>"u"?(s(),g(K,{key:0})):(s(),u(y,{key:1},[o(B,{type:"passthrough"},{default:t(()=>[o(R,{protocol:"passthrough",traffic:b.passthrough},{default:t(()=>e[24]||(e[24]=[a(`
                        Non mesh traffic
                      `)])),_:2},1032,["traffic"])]),_:2},1024),e[25]||(e[25]=a()),(s(),u(y,null,_(["upstream"],n=>o(D,{key:n,type:"outbounds",predicate:c.params.inactive?void 0:([p,l])=>{var f,k;return((typeof l.tcp<"u"?(f=l.tcp)==null?void 0:f[`${n}_cx_rx_bytes_total`]:(k=l.http)==null?void 0:k[`${n}_rq_total`])??0)>0},items:Object.entries(b.outbounds)},{default:t(({items:p})=>[p.length>0?(s(),g(B,{key:0,type:"outbound","data-testid":"dataplane-outbounds"},{default:t(()=>[(s(),u(y,null,_([/-([a-f0-9]){16}$/],l=>o(X,{key:l,type:"stack",size:"small"},{default:t(()=>[(s(!0),u(y,null,_(p,([f,k])=>(s(),g(R,{key:`${f}`,"data-testid":"dataplane-outbound",protocol:["grpc","http","tcp"].find(I=>typeof k[I]<"u")??"tcp",traffic:k,service:k.$resourceMeta.type===""?f.replace(l,""):void 0,direction:n},{default:t(()=>[o(z,{"data-action":"",to:{name:(I=>I.includes("bound")?I.replace("-inbound-","-outbound-"):"data-plane-connection-outbound-summary-overview-view")(String(S(V).name)),params:{connection:f},query:{inactive:c.params.inactive}}},{default:t(()=>[a(d(f),1)]),_:2},1032,["to"])]),_:2},1032,["protocol","traffic","service","direction"]))),128))]),_:2},1024)),64))]),_:2},1024)):C("",!0)]),_:2},1032,["predicate","items"])),64))],64))],64)):(s(),g(P,{key:1}))]),_:2},[b?{name:"actions",fn:t(()=>[o(H,{checked:c.params.inactive,"data-testid":"dataplane-outbounds-inactive-toggle",onChange:n=>c.update({inactive:n})},{label:t(()=>e[19]||(e[19]=[a(`
                      Show inactive
                    `)])),_:2},1032,["checked","onChange"]),e[21]||(e[21]=a()),o(z,{action:"refresh",appearance:"primary",onClick:Z},{default:t(()=>e[20]||(e[20]=[a(`
                    Refresh
                  `)])),_:2},1032,["onClick"])]),key:"0"}:void 0]),1024)]),_:2},1024)]),_:2},1024),e[45]||(e[45]=a()),o(L,null,{default:t(n=>[n.route.name!==c.name?(s(),g(rt,{key:0,width:"670px",onClose:function(){c.replace({name:"data-plane-detail-view",params:{mesh:c.params.mesh,proxy:c.params.proxy},query:{inactive:c.params.inactive?null:void 0}})}},{default:t(()=>[(s(),g(st(n.Component),{data:c.params.subscription.length>0?i.data.dataplaneInsight.subscriptions:n.route.name.includes("-inbound-")?i.data.dataplane.networking.inbounds:(b==null?void 0:b.outbounds)||{},networking:i.data.dataplane.networking},null,8,["data","networking"]))]),_:2},1032,["onClose"])):C("",!0)]),_:2},1024),e[46]||(e[46]=a()),w("div",yt,[w("h2",null,d(r("data-planes.routes.item.mtls.title")),1),e[38]||(e[38]=a()),i.data.dataplaneInsight.mTLS?(s(!0),u(y,{key:0},_([i.data.dataplaneInsight.mTLS],n=>(s(),g(q,{key:n,class:"mt-4"},{default:t(()=>[w("div",ct,[o(v,null,{title:t(()=>[a(d(r("data-planes.routes.item.mtls.expiration_time.title")),1)]),body:t(()=>[a(d(r("common.formats.datetime",{value:Date.parse(n.certificateExpirationTime)})),1)]),_:2},1024),e[34]||(e[34]=a()),o(v,null,{title:t(()=>[a(d(r("data-planes.routes.item.mtls.generation_time.title")),1)]),body:t(()=>[a(d(r("common.formats.datetime",{value:Date.parse(n.lastCertificateRegeneration)})),1)]),_:2},1024),e[35]||(e[35]=a()),o(v,null,{title:t(()=>[a(d(r("data-planes.routes.item.mtls.regenerations.title")),1)]),body:t(()=>[a(d(r("common.formats.integer",{value:n.certificateRegenerations})),1)]),_:2},1024),e[36]||(e[36]=a()),o(v,null,{title:t(()=>[a(d(r("data-planes.routes.item.mtls.issued_backend.title")),1)]),body:t(()=>[a(d(n.issuedBackend),1)]),_:2},1024),e[37]||(e[37]=a()),o(v,null,{title:t(()=>[a(d(r("data-planes.routes.item.mtls.supported_backends.title")),1)]),body:t(()=>[w("ul",null,[(s(!0),u(y,null,_(n.supportedBackends,p=>(s(),u("li",{key:p},d(p),1))),128))])]),_:2},1024)])]),_:2},1024))),128)):(s(),g(U,{key:1,class:"mt-4",variant:"warning"},{default:t(()=>[o(N,{path:"data-planes.routes.item.mtls.disabled"})]),_:1}))]),e[47]||(e[47]=a()),i.data.dataplaneInsight.subscriptions.length>0?(s(),u("div",gt,[w("h2",null,d(r("data-planes.routes.item.subscriptions.title")),1),e[43]||(e[43]=a()),o(dt,{headers:[{...x.get("headers.instanceId"),label:r("http.api.property.instanceId"),key:"instanceId"},{...x.get("headers.version"),label:r("http.api.property.version"),key:"version"},{...x.get("headers.connected"),label:r("http.api.property.connected"),key:"connected"},{...x.get("headers.disconnected"),label:r("http.api.property.disconnected"),key:"disconnected"},{...x.get("headers.responses"),label:r("http.api.property.responses"),key:"responses"}],"is-selected-row":n=>n.id===c.params.subscription,items:i.data.dataplaneInsight.subscriptions.map((n,p,l)=>l[l.length-(p+1)]),onResize:x.set},{instanceId:t(({row:n})=>[o(z,{"data-action":"",to:{name:"data-plane-subscription-summary-view",params:{subscription:n.id}}},{default:t(()=>[a(d(n.controlPlaneInstanceId),1)]),_:2},1032,["to"])]),version:t(({row:n})=>{var p,l;return[a(d(((l=(p=n.version)==null?void 0:p.kumaDp)==null?void 0:l.version)??"-"),1)]}),connected:t(({row:n})=>[a(d(r("common.formats.datetime",{value:Date.parse(n.connectTime??"")})),1)]),disconnected:t(({row:n})=>[n.disconnectTime?(s(),u(y,{key:0},[a(d(r("common.formats.datetime",{value:Date.parse(n.disconnectTime)})),1)],64)):C("",!0)]),responses:t(({row:n})=>{var p;return[(s(!0),u(y,null,_([((p=n.status)==null?void 0:p.total)??{}],l=>(s(),u(y,null,[a(d(l.responsesSent)+"/"+d(l.responsesAcknowledged),1)],64))),256))]}),_:2},1032,["headers","is-selected-row","items","onResize"])])):C("",!0)]),_:2},1024)]),_:2},[E.value.length>0||T?{name:"notifications",fn:t(()=>[w("ul",pt,[(s(!0),u(y,null,_(E.value,n=>(s(),u("li",{key:n.kind,"data-testid":`warning-${n.kind}`},[o(N,{path:`common.warnings.${n.kind}`,params:n.payload},null,8,["path","params"])],8,ut))),128)),e[2]||(e[2]=a()),T?(s(),u("li",mt,[e[0]||(e[0]=a(`
              The below view is not enhanced with runtime stats (Error loading stats: `)),w("strong",null,d(T.toString()),1),e[1]||(e[1]=a(`)
            `))])):C("",!0)])]),key:"0"}:void 0]),1024)]),_:2},1032,["src"])]),_:1})}}}),Ct=it(ft,[["__scopeId","data-v-85abb806"]]);export{Ct as default};
