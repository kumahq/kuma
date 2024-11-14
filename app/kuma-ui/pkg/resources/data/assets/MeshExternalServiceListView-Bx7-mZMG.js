import{d as E,e as a,o as t,m as u,w as o,a as r,b as l,l as k,ar as K,c as _,H as h,J as y,Q as X,ap as $,A as H,$ as M,t as d,p as f,E as P,k as q}from"./index-B_icS-nL.js";import{S as G}from"./SummaryView-DoPm-AWH.js";const F=["innerHTML"],j=E({__name:"MeshExternalServiceListView",props:{mesh:{}},setup(w){const x=w;return(I,m)=>{const b=a("RouteTitle"),v=a("XAction"),C=a("KumaPort"),V=a("XActionGroup"),A=a("RouterView"),D=a("DataCollection"),R=a("DataLoader"),S=a("KCard"),L=a("AppView"),N=a("DataSource"),B=a("RouteView");return t(),u(B,{name:"mesh-external-service-list-view",params:{page:1,size:50,mesh:"",service:""}},{default:o(({route:n,t:c,can:T,uri:z,me:p})=>[r(b,{render:!1,title:c("services.routes.mesh-external-service-list-view.title")},null,8,["title"]),m[6]||(m[6]=l()),r(N,{src:z(k(K),"/zone-cps/:name/egresses",{name:"*"},{page:1,size:100})},{default:o(({data:g})=>[(t(!0),_(h,null,y([[[c("services.mesh-external-service.notifications.mtls-warning"),typeof x.mesh.mtlsBackend>"u"],[c("services.mesh-external-service.notifications.no-zone-egress"),g&&!g.items.find(i=>typeof i.zoneEgressInsight.connectedSubscription<"u")]].filter(([i,s])=>s).map(i=>i[0])],i=>(t(),u(L,{key:typeof i,docs:c("services.mesh-external-service.href.docs")},X({default:o(()=>[m[5]||(m[5]=l()),r(S,null,{default:o(()=>[r(R,{src:z(k($),"/meshes/:mesh/mesh-external-services",{mesh:n.params.mesh},{page:n.params.page,size:n.params.size})},{loadable:o(({data:s})=>[r(D,{type:"services",items:(s==null?void 0:s.items)??[void 0],page:n.params.page,"page-size":n.params.size,total:s==null?void 0:s.total,onChange:n.update},{default:o(()=>[r(H,{"data-testid":"service-collection",headers:[{...p.get("headers.name"),label:"Name",key:"name"},{...p.get("headers.namespace"),label:"Namespace",key:"namespace"},...T("use zones")?[{...p.get("headers.zone"),label:"Zone",key:"zone"}]:[],{...p.get("headers.port"),label:"Port",key:"port"},{...p.get("headers.actions"),label:"Actions",key:"actions",hideLabel:!0}],items:s==null?void 0:s.items,"is-selected-row":e=>e.name===n.params.service,onResize:p.set},{name:o(({row:e})=>[r(M,{text:e.name},{default:o(()=>[r(v,{"data-action":"",to:{name:"mesh-external-service-summary-view",params:{mesh:e.mesh,service:e.id},query:{page:n.params.page,size:n.params.size}}},{default:o(()=>[l(d(e.name),1)]),_:2},1032,["to"])]),_:2},1032,["text"])]),namespace:o(({row:e})=>[l(d(e.namespace),1)]),zone:o(({row:e})=>[e.labels&&e.labels["kuma.io/origin"]==="zone"&&e.labels["kuma.io/zone"]?(t(),_(h,{key:0},[e.labels["kuma.io/zone"]?(t(),u(v,{key:0,to:{name:"zone-cp-detail-view",params:{zone:e.labels["kuma.io/zone"]}}},{default:o(()=>[l(d(e.labels["kuma.io/zone"]),1)]),_:2},1032,["to"])):f("",!0)],64)):(t(),_(h,{key:1},[l(d(c("common.detail.none")),1)],64))]),port:o(({row:e})=>[e.spec.match?(t(),u(C,{key:0,port:e.spec.match},null,8,["port"])):f("",!0)]),actions:o(({row:e})=>[r(V,null,{default:o(()=>[r(v,{to:{name:"mesh-external-service-detail-view",params:{mesh:e.mesh,service:e.id}}},{default:o(()=>[l(d(c("common.collection.actions.view")),1)]),_:2},1032,["to"])]),_:2},1024)]),_:2},1032,["headers","items","is-selected-row","onResize"]),m[4]||(m[4]=l()),s!=null&&s.items&&n.params.service?(t(),u(A,{key:0},{default:o(e=>[r(G,{onClose:J=>n.replace({name:"mesh-external-service-list-view",params:{mesh:n.params.mesh},query:{page:n.params.page,size:n.params.size}})},{default:o(()=>[(t(),u(P(e.Component),{items:s==null?void 0:s.items},null,8,["items"]))]),_:2},1032,["onClose"])]),_:2},1024)):f("",!0)]),_:2},1032,["items","page","page-size","total","onChange"])]),_:2},1032,["src"])]),_:2},1024)]),_:2},[i.length>0?{name:"notifications",fn:o(()=>[q("ul",null,[(t(!0),_(h,null,y(i,s=>(t(),_("li",{key:s,innerHTML:s},null,8,F))),128))])]),key:"0"}:void 0]),1032,["docs"]))),128))]),_:2},1032,["src"])]),_:1})}}});export{j as default};
