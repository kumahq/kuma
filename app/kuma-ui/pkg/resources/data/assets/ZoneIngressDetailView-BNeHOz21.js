import{d as F,r as p,o as s,p as m,w as e,b as a,Q as C,e as t,t as i,S as O,q as v,c,J as f,m as U,Z,K as z,l as E,A as G,F as J}from"./index-COAz_lIy.js";import{S as K}from"./SummaryView-Cu1duDPI.js";import{C as X,b as x,a as P}from"./ConnectionTraffic-CJOhIiJv.js";import"./TagList-DJqu--Rk.js";const Q={key:0},te=F({__name:"ZoneIngressDetailView",props:{data:{}},setup(D){const l=D;return(H,n)=>{const h=p("XBadge"),A=p("XCopyButton"),S=p("XAboutCard"),I=p("XLayout"),B=p("XInputSwitch"),V=p("XAction"),R=p("DataCollection"),T=p("XCard"),L=p("DataLoader"),$=p("RouterView"),N=p("AppView"),j=p("RouteView");return s(),m(j,{name:"zone-ingress-detail-view",params:{subscription:"",zoneIngress:"",inactive:!1}},{default:e(({t:d,me:_,route:y,uri:q})=>[a(N,null,{default:e(()=>[a(S,{title:d("zone-ingresses.routes.item.about.title"),created:l.data.creationTime,modified:l.data.modificationTime},{default:e(()=>[a(C,{layout:"horizontal"},{title:e(()=>[t(i(d("http.api.property.status")),1)]),body:e(()=>[a(O,{status:l.data.state},null,8,["status"])]),_:2},1024),n[4]||(n[4]=t()),l.data.namespace.length>0?(s(),m(C,{key:0,layout:"horizontal"},{title:e(()=>[t(i(d("http.api.property.namespace")),1)]),body:e(()=>[a(h,{appearance:"decorative"},{default:e(()=>[t(i(l.data.namespace),1)]),_:1})]),_:2},1024)):v("",!0),n[5]||(n[5]=t()),a(C,{layout:"horizontal"},{title:e(()=>[t(i(d("http.api.property.address")),1)]),body:e(()=>[l.data.zoneIngress.socketAddress.length>0?(s(),m(A,{key:0,variant:"badge",format:"default",text:l.data.zoneIngress.socketAddress},null,8,["text"])):(s(),c(f,{key:1},[t(i(d("common.detail.none")),1)],64))]),_:2},1024),n[6]||(n[6]=t()),a(C,{layout:"horizontal"},{title:e(()=>[t(i(d("http.api.property.advertisedAddress")),1)]),body:e(()=>[l.data.zoneIngress.advertisedSocketAddress.length>0?(s(),m(h,{key:0,appearance:"decorative"},{default:e(()=>[a(A,{text:l.data.zoneIngress.advertisedSocketAddress},null,8,["text"])]),_:1})):(s(),c(f,{key:1},[t(i(d("common.detail.none")),1)],64))]),_:2},1024)]),_:2},1032,["title","created","modified"]),n[18]||(n[18]=t()),a(L,{src:q(U(Z),"/connections/stats/for/zone-ingress/:name/:socketAddress",{name:y.params.zoneIngress,socketAddress:l.data.zoneIngress.socketAddress})},{default:e(({data:o,refresh:u})=>[a(T,{class:"traffic"},{default:e(()=>[a(I,{type:"columns"},{default:e(()=>[a(X,null,{default:e(()=>[a(I,{type:"stack",size:"small"},{default:e(()=>[(s(!0),c(f,null,z(Object.entries(o.inbounds),([r,g])=>(s(),m(x,{key:`${r}`,protocol:"",traffic:g},{default:e(()=>[t(`
                    :`+i(r.split("_").at(-1)),1)]),_:2},1032,["traffic"]))),128))]),_:2},1024)]),_:2},1024),n[11]||(n[11]=t()),a(X,null,{actions:e(()=>[a(B,{modelValue:y.params.inactive,"onUpdate:modelValue":r=>y.params.inactive=r,"data-testid":"dataplane-outbounds-inactive-toggle"},{label:e(()=>n[7]||(n[7]=[t(`
                    Show inactive
                  `)])),_:2},1032,["modelValue","onUpdate:modelValue"]),n[9]||(n[9]=t()),a(V,{action:"refresh",appearance:"primary",onClick:u},{default:e(()=>n[8]||(n[8]=[t(`
                  Refresh
                `)])),_:2},1032,["onClick"])]),default:e(()=>[n[10]||(n[10]=t()),(s(),c(f,null,z(["upstream"],r=>a(R,{key:r,predicate:y.params.inactive?void 0:([g,k])=>{var b,w;return((typeof k.tcp<"u"?(b=k.tcp)==null?void 0:b[`${r}_cx_rx_bytes_total`]:(w=k.http)==null?void 0:w[`${r}_rq_total`])??0)>0},items:Object.entries(o.outbounds)},{default:e(({items:g})=>[g.length>0?(s(),m(P,{key:0,type:"outbound"},{default:e(()=>[(s(),c(f,null,z([/-([a-f0-9]){16}$/],k=>a(I,{key:k,type:"stack",size:"small"},{default:e(()=>[(s(!0),c(f,null,z(g,([b,w])=>(s(),m(x,{key:`${b}`,"data-testid":"dataplane-outbound",protocol:"",traffic:w,direction:r},{default:e(()=>[t(i(b),1)]),_:2},1032,["traffic","direction"]))),128))]),_:2},1024)),64))]),_:2},1024)):v("",!0)]),_:2},1032,["predicate","items"])),64))]),_:2},1024)]),_:2},1024)]),_:2},1024)]),_:2},1032,["src"]),n[19]||(n[19]=t()),l.data.zoneIngressInsight.subscriptions.length>0?(s(),c("div",Q,[E("h2",null,i(d("zone-ingresses.routes.item.subscriptions.title")),1),n[16]||(n[16]=t()),a(G,{headers:[{..._.get("headers.instanceId"),label:d("http.api.property.instanceId"),key:"instanceId"},{..._.get("headers.version"),label:d("http.api.property.version"),key:"version"},{..._.get("headers.connected"),label:d("http.api.property.connected"),key:"connected"},{..._.get("headers.disconnected"),label:d("http.api.property.disconnected"),key:"disconnected"},{..._.get("headers.responses"),label:d("http.api.property.responses"),key:"responses"}],"is-selected-row":o=>o.id===y.params.subscription,items:l.data.zoneIngressInsight.subscriptions.map((o,u,r)=>r[r.length-(u+1)]),onResize:_.set},{instanceId:e(({row:o})=>[a(V,{"data-action":"",to:{name:"zone-ingress-subscription-summary-view",params:{subscription:o.id}}},{default:e(()=>[t(i(o.controlPlaneInstanceId),1)]),_:2},1032,["to"])]),version:e(({row:o})=>{var u,r;return[t(i(((r=(u=o.version)==null?void 0:u.kumaDp)==null?void 0:r.version)??"-"),1)]}),connected:e(({row:o})=>[t(i(d("common.formats.datetime",{value:Date.parse(o.connectTime??"")})),1)]),disconnected:e(({row:o})=>[o.disconnectTime?(s(),c(f,{key:0},[t(i(d("common.formats.datetime",{value:Date.parse(o.disconnectTime)})),1)],64)):v("",!0)]),responses:e(({row:o})=>{var u;return[(s(!0),c(f,null,z([((u=o.status)==null?void 0:u.total)??{}],r=>(s(),c(f,null,[t(i(r.responsesSent)+"/"+i(r.responsesAcknowledged),1)],64))),256))]}),_:2},1032,["headers","is-selected-row","items","onResize"]),n[17]||(n[17]=t()),a($,null,{default:e(({Component:o})=>[y.child()?(s(),m(K,{key:0,width:"670px",onClose:function(){y.replace({name:"zone-ingress-detail-view",params:{zoneIngress:y.params.zoneIngress}})}},{default:e(()=>[(s(),m(J(o),{data:l.data.zoneIngressInsight.subscriptions},null,8,["data"]))]),_:2},1032,["onClose"])):v("",!0)]),_:2},1024)])):v("",!0)]),_:2},1024)]),_:1})}}});export{te as default};
