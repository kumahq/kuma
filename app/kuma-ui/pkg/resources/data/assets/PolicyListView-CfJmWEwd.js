import{d as L,r as c,o as r,m as d,w as e,b as t,e as n,t as i,s as u,q as h,Z as S,p as F,a1 as q,C as E,c as k,F as C,K as $,_ as G}from"./index-C-Llvxgw.js";import{P as M}from"./PolicyTypeTag-BIjEzy0M.js";import{S as Z}from"./SummaryView-CzmbSpU2.js";const K={class:"stack"},j={class:"visually-hidden"},H={key:0},J=L({__name:"PolicyListView",props:{policyTypes:{}},setup(w){const X=w;return(O,o)=>{const z=c("XEmptyState"),_=c("XBadge"),f=c("XAction"),R=c("XI18n"),b=c("XCard"),V=c("XInput"),P=c("XIcon"),T=c("XActionGroup"),v=c("DataCollection"),x=c("RouterView"),A=c("DataLoader"),I=c("AppView"),N=c("RouteView");return r(),d(N,{name:"policy-list-view",params:{page:1,size:Number,mesh:"",policyPath:"",policy:"",s:""}},{default:e(({route:s,t:p,can:B,uri:D,me:y})=>[t(v,{predicate:l=>typeof l<"u"&&l.path===s.params.policyPath,items:X.policyTypes??[]},{empty:e(()=>[t(z,null,{default:e(()=>[n(i(p("policies.routes.items.empty")),1)]),_:2},1024)]),item:e(({item:l})=>[t(I,null,{default:e(()=>[u("div",K,[t(b,null,{default:e(()=>[u("header",null,[u("div",null,[l.policy.hasFromTargetRef?(r(),d(_,{key:0,appearance:"neutral"},{default:e(()=>[n(i(p("policies.collection.inbound")),1)]),_:2},1024)):h("",!0),o[1]||(o[1]=n()),l.policy.hasToTargetRef?(r(),d(_,{key:1,appearance:"neutral"},{default:e(()=>[n(i(p("policies.collection.outbound")),1)]),_:2},1024)):h("",!0),o[2]||(o[2]=n()),t(f,{action:"docs",href:p("policies.href.docs",{name:l.name}),"data-testid":"policy-documentation-link"},{default:e(()=>[u("span",j,i(p("common.documentation")),1)]),_:2},1032,["href"])]),o[3]||(o[3]=n()),u("h3",null,[t(M,{"policy-type":l.name},{default:e(()=>[n(i(p("policies.collection.title",{name:l.name})),1)]),_:2},1032,["policy-type"])])]),o[4]||(o[4]=n()),t(R,{path:`policies.type.${l.name}.description`,"default-path":"policies.collection.description"},null,8,["path"])]),_:2},1024),o[17]||(o[17]=n()),t(b,null,{default:e(()=>[u("search",null,[u("form",{onSubmit:o[0]||(o[0]=S(()=>{},["prevent"]))},[t(V,{placeholder:"Filter by name...",type:"search",appearance:"search",value:s.params.s,debounce:1e3,onChange:m=>s.update({s:m})},null,8,["value","onChange"])],32)]),o[16]||(o[16]=n()),t(A,{src:D(F(q),"/meshes/:mesh/policy-path/:path",{mesh:s.params.mesh,path:s.params.policyPath},{page:s.params.page,size:s.params.size,search:s.params.s})},{loadable:e(({data:m})=>[t(v,{items:(m==null?void 0:m.items)??[void 0],page:s.params.page,"page-size":s.params.size,total:m==null?void 0:m.total,onChange:s.update},{empty:e(()=>[t(z,null,{title:e(()=>[u("h3",null,i(p("policies.x-empty-state.title")),1)]),action:e(()=>[t(f,{action:"docs",href:p("policies.href.docs",{name:l.name})},{default:e(()=>[n(i(p("common.documentation")),1)]),_:2},1032,["href"])]),default:e(()=>[o[5]||(o[5]=n()),t(R,{path:"policies.x-empty-state.body",params:{type:l.name,suffix:s.params.s.length>0?p("common.matchingsearch"):""}},null,8,["params"]),o[6]||(o[6]=n())]),_:2},1024)]),default:e(()=>[t(E,{headers:[{...y.get("headers.role"),label:"Role",key:"role",hideLabel:!0},{...y.get("headers.name"),label:"Name",key:"name"},{...y.get("headers.namespace"),label:"Namespace",key:"namespace"},...B("use zones")&&l.policy.isTargetRef?[{...y.get("headers.zone"),label:"Zone",key:"zone"}]:[],...l.policy.isTargetRef?[{...y.get("headers.targetRef"),label:"Target ref",key:"targetRef"}]:[],{...y.get("headers.actions"),label:"Actions",key:"actions",hideLabel:!0}],items:m==null?void 0:m.items,"is-selected-row":a=>a.id===s.params.policy,onResize:y.set},{role:e(({row:a})=>[a.role==="producer"?(r(),d(P,{key:0,name:`policy-role-${a.role}`},{default:e(()=>[n(`
                              Role: `+i(a.role),1)]),_:2},1032,["name"])):(r(),k(C,{key:1},[n(`
                             
                          `)],64))]),name:e(({row:a})=>[t(f,{"data-action":"",to:{name:"policy-summary-view",params:{mesh:a.mesh,policyPath:l.path,policy:a.id},query:{page:s.params.page,size:s.params.size}}},{default:e(()=>[n(i(a.name),1)]),_:2},1032,["to"])]),namespace:e(({row:a})=>[n(i(a.namespace.length>0?a.namespace:p("common.detail.none")),1)]),targetRef:e(({row:a})=>{var g;return[typeof((g=a.spec)==null?void 0:g.targetRef)<"u"?(r(),d(_,{key:0,appearance:"neutral"},{default:e(()=>[n(i(a.spec.targetRef.kind),1),a.spec.targetRef.name?(r(),k("span",H,[o[7]||(o[7]=n(":")),u("b",null,i(a.spec.targetRef.name),1)])):h("",!0)]),_:2},1024)):(r(),d(_,{key:1,appearance:"neutral"},{default:e(()=>o[8]||(o[8]=[n(`
                            Mesh
                          `)])),_:1}))]}),zone:e(({row:a})=>[a.zone?(r(),d(f,{key:0,to:{name:"zone-cp-detail-view",params:{zone:a.zone}}},{default:e(()=>[n(i(a.zone),1)]),_:2},1032,["to"])):(r(),k(C,{key:1},[n(i(p("common.detail.none")),1)],64))]),actions:e(({row:a})=>[t(T,null,{default:e(()=>[t(f,{to:{name:"policy-detail-view",params:{mesh:a.mesh,policyPath:l.path,policy:a.id}}},{default:e(()=>[n(i(p("common.collection.actions.view")),1)]),_:2},1032,["to"])]),_:2},1024)]),_:2},1032,["headers","items","is-selected-row","onResize"])]),_:2},1032,["items","page","page-size","total","onChange"]),o[15]||(o[15]=n()),s.params.policy?(r(),d(x,{key:0},{default:e(({Component:a})=>[t(Z,{onClose:g=>s.replace({name:"policy-list-view",params:{mesh:s.params.mesh,policyPath:s.params.policyPath},query:{page:s.params.page,size:s.params.size}})},{default:e(()=>[typeof m<"u"?(r(),d($(a),{key:0,items:m.items,"policy-type":l},null,8,["items","policy-type"])):h("",!0)]),_:2},1032,["onClose"])]),_:2},1024)):h("",!0)]),_:2},1032,["src"])]),_:2},1024)])]),_:2},1024)]),_:2},1032,["predicate","items"])]),_:1})}}}),Y=G(J,[["__scopeId","data-v-a580cc30"]]);export{Y as default};
