import{d as N,r as p,o as c,m,w as e,b as n,k as u,Z as _,e as a,t as l,p as r,c as w,l as B,ay as L,A as x,L as K,E as X}from"./index-Di1IipgC.js";import{S}from"./SummaryView-CR0LPp2b.js";const Z={class:"columns"},E={key:0},G=u("h3",null,`
          Affected Data Plane Proxies
        `,-1),j=N({__name:"PolicyDetailView",props:{data:{}},setup(C){const s=C;return(T,q)=>{const h=p("XAction"),f=p("KBadge"),g=p("KCard"),z=p("RouterLink"),R=p("XActionGroup"),b=p("RouterView"),v=p("DataCollection"),V=p("DataLoader"),A=p("AppView"),D=p("RouteView");return c(),m(D,{name:"policy-detail-view",params:{page:1,size:50,s:"",mesh:"",policy:"",policyPath:"",dataPlane:""}},{default:e(({route:t,t:y,uri:P,can:k,me:d})=>[n(A,null,{default:e(()=>[n(g,null,{default:e(()=>[u("div",Z,[n(_,null,{title:e(()=>[a(`
              Type
            `)]),body:e(()=>[a(l(s.data.type),1)]),_:1}),a(),s.data.namespace.length>0?(c(),m(_,{key:0},{title:e(()=>[a(`
              Namespace
            `)]),body:e(()=>[a(l(s.data.namespace),1)]),_:1})):r("",!0),a(),k("use zones")&&s.data.zone?(c(),m(_,{key:1},{title:e(()=>[a(`
              Zone
            `)]),body:e(()=>[n(h,{to:{name:"zone-cp-detail-view",params:{zone:s.data.zone}}},{default:e(()=>[a(l(s.data.zone),1)]),_:1},8,["to"])]),_:1})):r("",!0),a(),s.data.spec?(c(),m(_,{key:2},{title:e(()=>[a(l(y("http.api.property.targetRef")),1)]),body:e(()=>[s.data.spec.targetRef?(c(),m(f,{key:0,appearance:"neutral"},{default:e(()=>[a(l(s.data.spec.targetRef.kind),1),s.data.spec.targetRef.name?(c(),w("span",E,[a(":"),u("b",null,l(s.data.spec.targetRef.name),1)])):r("",!0)]),_:1})):(c(),m(f,{key:1,appearance:"neutral"},{default:e(()=>[a(`
                Mesh
              `)]),_:1}))]),_:2},1024)):r("",!0)])]),_:2},1024),a(),u("div",null,[G,a(),n(g,{class:"mt-4"},{default:e(()=>[n(V,{src:P(B(L),"/meshes/:mesh/policy-path/:path/policy/:name/dataplanes",{mesh:t.params.mesh,path:t.params.policyPath,name:t.params.policy},{page:t.params.page,size:t.params.size})},{loadable:e(({data:i})=>[n(v,{type:"data-planes",items:(i==null?void 0:i.items)??[void 0],page:t.params.page,"page-size":t.params.size,total:i==null?void 0:i.total,onChange:t.update},{default:e(()=>[n(x,{headers:[{...d.get("headers.name"),label:"Name",key:"name"},{...d.get("headers.namespace"),label:"Namespace",key:"namespace"},...k("use zones")?[{...d.get("headers.zone"),label:"Zone",key:"zone"}]:[],{...d.get("headers.actions"),label:"Actions",key:"actions",hideLabel:!0}],items:i==null?void 0:i.items,"is-selected-row":o=>o.id===t.params.dataPlane,onResize:d.set},{name:e(({row:o})=>[n(z,{"data-action":"",to:{name:"data-plane-detail-view",params:{dataPlane:o.id}}},{default:e(()=>[a(l(o.name),1)]),_:2},1032,["to"])]),namespace:e(({row:o})=>[a(l(o.namespace),1)]),zone:e(({row:o})=>[o.zone?(c(),m(z,{key:0,to:{name:"zone-cp-detail-view",params:{zone:o.zone}}},{default:e(()=>[a(l(o.zone),1)]),_:2},1032,["to"])):(c(),w(K,{key:1},[a(l(y("common.collection.none")),1)],64))]),actions:e(({row:o})=>[n(R,null,{default:e(()=>[n(h,{to:{name:"data-plane-detail-view",params:{dataPlane:o.id}}},{default:e(()=>[a(l(y("common.collection.actions.view")),1)]),_:2},1032,["to"])]),_:2},1024)]),_:2},1032,["headers","items","is-selected-row","onResize"]),a(),n(b,null,{default:e(({Component:o})=>[t.child()?(c(),m(S,{key:0,onClose:F=>t.replace({params:{mesh:t.params.mesh},query:{page:t.params.page,size:t.params.size,s:t.params.s}})},{default:e(()=>[typeof i<"u"?(c(),m(X(o),{key:0,items:i.items},null,8,["items"])):r("",!0)]),_:2},1032,["onClose"])):r("",!0)]),_:2},1024)]),_:2},1032,["items","page","page-size","total","onChange"])]),_:2},1032,["src"])]),_:2},1024)])]),_:2},1024)]),_:1})}}});export{j as default};
