import{d as A,r as s,o as i,m as c,w as a,b as t,l as b,aA as v,A as R,e as l,t as m,c as D,F as P,E as L,p as d}from"./index-DHg9Fngg.js";import{S as N}from"./SummaryView-DmzBN36G.js";const G=A({__name:"PolicyDetailView",setup(x){return(X,B)=>{const r=s("RouterLink"),u=s("XAction"),h=s("XActionGroup"),y=s("DataCollection"),z=s("RouterView"),f=s("DataLoader"),g=s("KCard"),w=s("AppView"),C=s("RouteView");return i(),c(C,{name:"policy-detail-view",params:{page:1,size:50,s:"",mesh:"",policy:"",policyPath:"",dataPlane:""}},{default:a(({route:e,t:_,uri:k,can:V,me:p})=>[t(w,null,{default:a(()=>[t(g,null,{default:a(()=>[t(f,{src:k(b(v),"/meshes/:mesh/policy-path/:path/policy/:name/dataplanes",{mesh:e.params.mesh,path:e.params.policyPath,name:e.params.policy},{page:e.params.page,size:e.params.size})},{loadable:a(({data:o})=>[t(y,{type:"data-planes",items:(o==null?void 0:o.items)??[void 0]},{default:a(()=>[t(R,{"page-number":e.params.page,"page-size":e.params.size,headers:[{...p.get("headers.name"),label:"Name",key:"name"},{...p.get("headers.namespace"),label:"Namespace",key:"namespace"},...V("use zones")?[{...p.get("headers.zone"),label:"Zone",key:"zone"}]:[],{...p.get("headers.actions"),label:"Actions",key:"actions",hideLabel:!0}],items:o==null?void 0:o.items,total:o==null?void 0:o.total,"is-selected-row":n=>n.id===e.params.dataPlane,onChange:e.update,onResize:p.set},{name:a(({row:n})=>[t(r,{"data-action":"",to:{name:"data-plane-detail-view",params:{dataPlane:n.id}}},{default:a(()=>[l(m(n.name),1)]),_:2},1032,["to"])]),namespace:a(({row:n})=>[l(m(n.namespace),1)]),zone:a(({row:n})=>[n.zone?(i(),c(r,{key:0,to:{name:"zone-cp-detail-view",params:{zone:n.zone}}},{default:a(()=>[l(m(n.zone),1)]),_:2},1032,["to"])):(i(),D(P,{key:1},[l(m(_("common.collection.none")),1)],64))]),actions:a(({row:n})=>[t(h,null,{default:a(()=>[t(u,{to:{name:"data-plane-detail-view",params:{dataPlane:n.id}}},{default:a(()=>[l(m(_("common.collection.actions.view")),1)]),_:2},1032,["to"])]),_:2},1024)]),_:2},1032,["page-number","page-size","headers","items","total","is-selected-row","onChange","onResize"])]),_:2},1032,["items"]),l(),t(z,null,{default:a(({Component:n})=>[e.child()?(i(),c(N,{key:0,onClose:S=>e.replace({params:{mesh:e.params.mesh},query:{page:e.params.page,size:e.params.size,s:e.params.s}})},{default:a(()=>[typeof o<"u"?(i(),c(L(n),{key:0,items:o.items},null,8,["items"])):d("",!0)]),_:2},1032,["onClose"])):d("",!0)]),_:2},1024)]),_:2},1032,["src"])]),_:2},1024)]),_:2},1024)]),_:1})}}});export{G as default};
