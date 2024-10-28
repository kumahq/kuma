import{d as D,e as m,o as p,m as r,w as e,a as i,b as n,t as l,k as d,p as _,l as I,aq as N,c as f,al as S,A as E,H as w,E as H,q}from"./index-C4IVBmnO.js";import{P as $}from"./PolicyTypeTag-BWgvkWVS.js";import{S as F}from"./SummaryView-B2TyI4rm.js";const G={class:"stack"},K={class:"visually-hidden"},O=["innerHTML"],Z={key:0},j=["innerHTML"],J={key:0},Q=D({__name:"PolicyListView",props:{policyTypes:{}},setup(R){const C=R;return(U,k)=>{const z=m("XEmptyState"),h=m("XBadge"),y=m("XAction"),b=m("KCard"),V=m("XInput"),X=m("XIcon"),P=m("XActionGroup"),v=m("DataCollection"),T=m("RouterView"),L=m("DataLoader"),x=m("AppView"),A=m("RouteView");return p(),r(A,{name:"policy-list-view",params:{page:1,size:50,mesh:"",policyPath:"",policy:"",s:""}},{default:e(({route:o,t:s,can:B,uri:M,me:u})=>[i(v,{predicate:t=>typeof t<"u"&&t.path===o.params.policyPath,items:C.policyTypes??[]},{empty:e(()=>[i(z,null,{default:e(()=>[n(l(s("policies.routes.items.empty")),1)]),_:2},1024)]),item:e(({item:t})=>[i(x,null,{default:e(()=>[d("div",G,[i(b,null,{default:e(()=>[d("header",null,[d("div",null,[t.isExperimental?(p(),r(h,{key:0,appearance:"warning"},{default:e(()=>[n(l(s("policies.collection.beta")),1)]),_:2},1024)):_("",!0),n(),t.isInbound?(p(),r(h,{key:1,appearance:"neutral"},{default:e(()=>[n(l(s("policies.collection.inbound")),1)]),_:2},1024)):_("",!0),n(),t.isOutbound?(p(),r(h,{key:2,appearance:"neutral"},{default:e(()=>[n(l(s("policies.collection.outbound")),1)]),_:2},1024)):_("",!0),n(),i(y,{action:"docs",href:s("policies.href.docs",{name:t.name}),"data-testid":"policy-documentation-link"},{default:e(()=>[d("span",K,l(s("common.documentation")),1)]),_:2},1032,["href"])]),n(),d("h3",null,[i($,{"policy-type":t.name},{default:e(()=>[n(l(s("policies.collection.title",{name:t.name})),1)]),_:2},1032,["policy-type"])])]),n(),d("div",{innerHTML:s(`policies.type.${t.name}.description`,void 0,{defaultMessage:s("policies.collection.description")})},null,8,O)]),_:2},1024),n(),i(b,null,{default:e(()=>[i(L,{src:M(I(N),"/meshes/:mesh/policy-path/:path",{mesh:o.params.mesh,path:o.params.policyPath},{page:o.params.page,size:o.params.size,search:o.params.s})},{loadable:e(({data:c})=>[((c==null?void 0:c.items)??{length:0}).length>0||o.params.s.length>0?(p(),f("search",Z,[d("form",{onSubmit:k[0]||(k[0]=S(()=>{},["prevent"]))},[i(V,{placeholder:"Filter by name...",type:"search",appearance:"search",value:o.params.s,debounce:1e3,onChange:a=>o.update({s:a})},null,8,["value","onChange"])],32)])):_("",!0),n(),i(v,{items:(c==null?void 0:c.items)??[void 0],page:o.params.page,"page-size":o.params.size,total:c==null?void 0:c.total,onChange:o.update},{empty:e(()=>[i(z,null,{title:e(()=>[d("h3",null,l(s("policies.x-empty-state.title")),1)]),action:e(()=>[i(y,{action:"docs",href:s("policies.href.docs",{name:t.name})},{default:e(()=>[n(l(s("common.documentation")),1)]),_:2},1032,["href"])]),default:e(()=>[n(),d("div",{innerHTML:s("policies.x-empty-state.body",{type:t.name,suffix:o.params.s.length>0?s("common.matchingsearch"):""})},null,8,j),n()]),_:2},1024)]),default:e(()=>[i(E,{headers:[{...u.get("headers.role"),label:"Role",key:"role",hideLabel:!0},{...u.get("headers.name"),label:"Name",key:"name"},{...u.get("headers.namespace"),label:"Namespace",key:"namespace"},...B("use zones")&&t.isTargetRefBased?[{...u.get("headers.zone"),label:"Zone",key:"zone"}]:[],...t.isTargetRefBased?[{...u.get("headers.targetRef"),label:"Target ref",key:"targetRef"}]:[],{...u.get("headers.actions"),label:"Actions",key:"actions",hideLabel:!0}],items:c==null?void 0:c.items,"is-selected-row":a=>a.id===o.params.policy,onResize:u.set},{role:e(({row:a})=>[a.role==="producer"?(p(),r(X,{key:0,name:`policy-role-${a.role}`},{default:e(()=>[n(`
                              Role: `+l(a.role),1)]),_:2},1032,["name"])):(p(),f(w,{key:1},[n(`
                             
                          `)],64))]),name:e(({row:a})=>[i(y,{"data-action":"",to:{name:"policy-summary-view",params:{mesh:a.mesh,policyPath:t.path,policy:a.id},query:{page:o.params.page,size:o.params.size}}},{default:e(()=>[n(l(a.name),1)]),_:2},1032,["to"])]),namespace:e(({row:a})=>[n(l(a.namespace.length>0?a.namespace:s("common.detail.none")),1)]),targetRef:e(({row:a})=>{var g;return[typeof((g=a.spec)==null?void 0:g.targetRef)<"u"?(p(),r(h,{key:0,appearance:"neutral"},{default:e(()=>[n(l(a.spec.targetRef.kind),1),a.spec.targetRef.name?(p(),f("span",J,[n(":"),d("b",null,l(a.spec.targetRef.name),1)])):_("",!0)]),_:2},1024)):(p(),r(h,{key:1,appearance:"neutral"},{default:e(()=>[n(`
                            Mesh
                          `)]),_:1}))]}),zone:e(({row:a})=>[a.zone?(p(),r(y,{key:0,to:{name:"zone-cp-detail-view",params:{zone:a.zone}}},{default:e(()=>[n(l(a.zone),1)]),_:2},1032,["to"])):(p(),f(w,{key:1},[n(l(s("common.detail.none")),1)],64))]),actions:e(({row:a})=>[i(P,null,{default:e(()=>[i(y,{to:{name:"policy-detail-view",params:{mesh:a.mesh,policyPath:t.path,policy:a.id}}},{default:e(()=>[n(l(s("common.collection.actions.view")),1)]),_:2},1032,["to"])]),_:2},1024)]),_:2},1032,["headers","items","is-selected-row","onResize"])]),_:2},1032,["items","page","page-size","total","onChange"]),n(),o.params.policy?(p(),r(T,{key:1},{default:e(({Component:a})=>[i(F,{onClose:g=>o.replace({name:"policy-list-view",params:{mesh:o.params.mesh,policyPath:o.params.policyPath},query:{page:o.params.page,size:o.params.size}})},{default:e(()=>[typeof c<"u"?(p(),r(H(a),{key:0,items:c.items,"policy-type":t},null,8,["items","policy-type"])):_("",!0)]),_:2},1032,["onClose"])]),_:2},1024)):_("",!0)]),_:2},1032,["src"])]),_:2},1024)])]),_:2},1024)]),_:2},1032,["predicate","items"])]),_:1})}}}),ae=q(Q,[["__scopeId","data-v-4f5d2159"]]);export{ae as default};
