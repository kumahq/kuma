import{d as B,e as n,o as t,m as p,w as o,a as r,b as c,l as g,au as E,c as u,H as h,J as k,T as K,as as X,A as H,a1 as M,t as _,p as v,E as P,k as $}from"./index-COT-_p62.js";import{S as q}from"./SummaryView-BaqCyaq5.js";const G=["innerHTML"],O=B({__name:"MeshExternalServiceListView",props:{mesh:{}},setup(y){const w=y;return(F,I)=>{const x=n("RouteTitle"),d=n("XAction"),b=n("KumaPort"),C=n("XActionGroup"),V=n("RouterView"),A=n("DataCollection"),D=n("DataLoader"),R=n("KCard"),S=n("AppView"),L=n("DataSource"),N=n("RouteView");return t(),p(N,{name:"mesh-external-service-list-view",params:{page:1,size:50,mesh:"",service:""}},{default:o(({route:a,t:l,can:T,uri:f,me:m})=>[r(x,{render:!1,title:l("services.routes.mesh-external-service-list-view.title")},null,8,["title"]),c(),r(L,{src:f(g(E),"/zone-cps/:name/egresses",{name:"*"},{page:1,size:100})},{default:o(({data:z})=>[(t(!0),u(h,null,k([[[l("services.mesh-external-service.notifications.mtls-warning"),typeof w.mesh.mtlsBackend>"u"],[l("services.mesh-external-service.notifications.no-zone-egress"),z&&!z.items.find(i=>typeof i.zoneEgressInsight.connectedSubscription<"u")]].filter(([i,s])=>s).map(i=>i[0])],i=>(t(),p(S,{key:typeof i,docs:l("services.mesh-external-service.href.docs")},K({default:o(()=>[c(),r(R,null,{default:o(()=>[r(D,{src:f(g(X),"/meshes/:mesh/mesh-external-services",{mesh:a.params.mesh},{page:a.params.page,size:a.params.size})},{loadable:o(({data:s})=>[r(A,{type:"services",items:(s==null?void 0:s.items)??[void 0],page:a.params.page,"page-size":a.params.size,total:s==null?void 0:s.total,onChange:a.update},{default:o(()=>[r(H,{"data-testid":"service-collection",headers:[{...m.get("headers.name"),label:"Name",key:"name"},{...m.get("headers.namespace"),label:"Namespace",key:"namespace"},...T("use zones")?[{...m.get("headers.zone"),label:"Zone",key:"zone"}]:[],{...m.get("headers.port"),label:"Port",key:"port"},{...m.get("headers.actions"),label:"Actions",key:"actions",hideLabel:!0}],items:s==null?void 0:s.items,"is-selected-row":e=>e.name===a.params.service,onResize:m.set},{name:o(({row:e})=>[r(M,{text:e.name},{default:o(()=>[r(d,{"data-action":"",to:{name:"mesh-external-service-summary-view",params:{mesh:e.mesh,service:e.id},query:{page:a.params.page,size:a.params.size}}},{default:o(()=>[c(_(e.name),1)]),_:2},1032,["to"])]),_:2},1032,["text"])]),namespace:o(({row:e})=>[c(_(e.namespace),1)]),zone:o(({row:e})=>[e.labels&&e.labels["kuma.io/origin"]==="zone"&&e.labels["kuma.io/zone"]?(t(),u(h,{key:0},[e.labels["kuma.io/zone"]?(t(),p(d,{key:0,to:{name:"zone-cp-detail-view",params:{zone:e.labels["kuma.io/zone"]}}},{default:o(()=>[c(_(e.labels["kuma.io/zone"]),1)]),_:2},1032,["to"])):v("",!0)],64)):(t(),u(h,{key:1},[c(_(l("common.detail.none")),1)],64))]),port:o(({row:e})=>[e.spec.match?(t(),p(b,{key:0,port:e.spec.match},null,8,["port"])):v("",!0)]),actions:o(({row:e})=>[r(C,null,{default:o(()=>[r(d,{to:{name:"mesh-external-service-detail-view",params:{mesh:e.mesh,service:e.id}}},{default:o(()=>[c(_(l("common.collection.actions.view")),1)]),_:2},1032,["to"])]),_:2},1024)]),_:2},1032,["headers","items","is-selected-row","onResize"]),c(),s!=null&&s.items&&a.params.service?(t(),p(V,{key:0},{default:o(e=>[r(q,{onClose:J=>a.replace({name:"mesh-external-service-list-view",params:{mesh:a.params.mesh},query:{page:a.params.page,size:a.params.size}})},{default:o(()=>[(t(),p(P(e.Component),{items:s==null?void 0:s.items},null,8,["items"]))]),_:2},1032,["onClose"])]),_:2},1024)):v("",!0)]),_:2},1032,["items","page","page-size","total","onChange"])]),_:2},1032,["src"])]),_:2},1024)]),_:2},[i.length>0?{name:"notifications",fn:o(()=>[$("ul",null,[(t(!0),u(h,null,k(i,s=>(t(),u("li",{key:s,innerHTML:s},null,8,G))),128))])]),key:"0"}:void 0]),1032,["docs"]))),128))]),_:2},1032,["src"])]),_:1})}}});export{O as default};
