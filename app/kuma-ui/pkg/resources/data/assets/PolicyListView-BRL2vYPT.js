import{d as D,e as r,o as c,m as u,w as e,a as l,b as o,t as p,k as d,p as f,am as I,l as N,ao as S,A as E,c as k,H as w,E as H,q}from"./index-C_eW3RRu.js";import{P as $}from"./PolicyTypeTag-BPWXGVTV.js";import{S as F}from"./SummaryView-C9wRmLik.js";const G={class:"stack"},K={class:"visually-hidden"},O=["innerHTML"],Z=["innerHTML"],j={key:0},J=D({__name:"PolicyListView",props:{policyTypes:{}},setup(R){const C=R;return(Q,a)=>{const z=r("XEmptyState"),_=r("XBadge"),g=r("XAction"),b=r("KCard"),V=r("XInput"),X=r("XIcon"),P=r("XActionGroup"),v=r("DataCollection"),T=r("RouterView"),L=r("DataLoader"),x=r("AppView"),A=r("RouteView");return c(),u(A,{name:"policy-list-view",params:{page:1,size:50,mesh:"",policyPath:"",policy:"",s:""}},{default:e(({route:s,t,can:B,uri:M,me:y})=>[l(v,{predicate:i=>typeof i<"u"&&i.path===s.params.policyPath,items:C.policyTypes??[]},{empty:e(()=>[l(z,null,{default:e(()=>[o(p(t("policies.routes.items.empty")),1)]),_:2},1024)]),item:e(({item:i})=>[l(x,null,{default:e(()=>[d("div",G,[l(b,null,{default:e(()=>[d("header",null,[d("div",null,[i.isExperimental?(c(),u(_,{key:0,appearance:"warning"},{default:e(()=>[o(p(t("policies.collection.beta")),1)]),_:2},1024)):f("",!0),a[1]||(a[1]=o()),i.isInbound?(c(),u(_,{key:1,appearance:"neutral"},{default:e(()=>[o(p(t("policies.collection.inbound")),1)]),_:2},1024)):f("",!0),a[2]||(a[2]=o()),i.isOutbound?(c(),u(_,{key:2,appearance:"neutral"},{default:e(()=>[o(p(t("policies.collection.outbound")),1)]),_:2},1024)):f("",!0),a[3]||(a[3]=o()),l(g,{action:"docs",href:t("policies.href.docs",{name:i.name}),"data-testid":"policy-documentation-link"},{default:e(()=>[d("span",K,p(t("common.documentation")),1)]),_:2},1032,["href"])]),a[4]||(a[4]=o()),d("h3",null,[l($,{"policy-type":i.name},{default:e(()=>[o(p(t("policies.collection.title",{name:i.name})),1)]),_:2},1032,["policy-type"])])]),a[5]||(a[5]=o()),d("div",{innerHTML:t(`policies.type.${i.name}.description`,void 0,{defaultMessage:t("policies.collection.description")})},null,8,O)]),_:2},1024),a[18]||(a[18]=o()),l(b,null,{default:e(()=>[d("search",null,[d("form",{onSubmit:a[0]||(a[0]=I(()=>{},["prevent"]))},[l(V,{placeholder:"Filter by name...",type:"search",appearance:"search",value:s.params.s,debounce:1e3,onChange:m=>s.update({s:m})},null,8,["value","onChange"])],32)]),a[17]||(a[17]=o()),l(L,{src:M(N(S),"/meshes/:mesh/policy-path/:path",{mesh:s.params.mesh,path:s.params.policyPath},{page:s.params.page,size:s.params.size,search:s.params.s})},{loadable:e(({data:m})=>[l(v,{items:(m==null?void 0:m.items)??[void 0],page:s.params.page,"page-size":s.params.size,total:m==null?void 0:m.total,onChange:s.update},{empty:e(()=>[l(z,null,{title:e(()=>[d("h3",null,p(t("policies.x-empty-state.title")),1)]),action:e(()=>[l(g,{action:"docs",href:t("policies.href.docs",{name:i.name})},{default:e(()=>[o(p(t("common.documentation")),1)]),_:2},1032,["href"])]),default:e(()=>[a[6]||(a[6]=o()),d("div",{innerHTML:t("policies.x-empty-state.body",{type:i.name,suffix:s.params.s.length>0?t("common.matchingsearch"):""})},null,8,Z),a[7]||(a[7]=o())]),_:2},1024)]),default:e(()=>[l(E,{headers:[{...y.get("headers.role"),label:"Role",key:"role",hideLabel:!0},{...y.get("headers.name"),label:"Name",key:"name"},{...y.get("headers.namespace"),label:"Namespace",key:"namespace"},...B("use zones")&&i.isTargetRefBased?[{...y.get("headers.zone"),label:"Zone",key:"zone"}]:[],...i.isTargetRefBased?[{...y.get("headers.targetRef"),label:"Target ref",key:"targetRef"}]:[],{...y.get("headers.actions"),label:"Actions",key:"actions",hideLabel:!0}],items:m==null?void 0:m.items,"is-selected-row":n=>n.id===s.params.policy,onResize:y.set},{role:e(({row:n})=>[n.role==="producer"?(c(),u(X,{key:0,name:`policy-role-${n.role}`},{default:e(()=>[o(`
                              Role: `+p(n.role),1)]),_:2},1032,["name"])):(c(),k(w,{key:1},[o(`
                             
                          `)],64))]),name:e(({row:n})=>[l(g,{"data-action":"",to:{name:"policy-summary-view",params:{mesh:n.mesh,policyPath:i.path,policy:n.id},query:{page:s.params.page,size:s.params.size}}},{default:e(()=>[o(p(n.name),1)]),_:2},1032,["to"])]),namespace:e(({row:n})=>[o(p(n.namespace.length>0?n.namespace:t("common.detail.none")),1)]),targetRef:e(({row:n})=>{var h;return[typeof((h=n.spec)==null?void 0:h.targetRef)<"u"?(c(),u(_,{key:0,appearance:"neutral"},{default:e(()=>[o(p(n.spec.targetRef.kind),1),n.spec.targetRef.name?(c(),k("span",j,[a[8]||(a[8]=o(":")),d("b",null,p(n.spec.targetRef.name),1)])):f("",!0)]),_:2},1024)):(c(),u(_,{key:1,appearance:"neutral"},{default:e(()=>a[9]||(a[9]=[o(`
                            Mesh
                          `)])),_:1}))]}),zone:e(({row:n})=>[n.zone?(c(),u(g,{key:0,to:{name:"zone-cp-detail-view",params:{zone:n.zone}}},{default:e(()=>[o(p(n.zone),1)]),_:2},1032,["to"])):(c(),k(w,{key:1},[o(p(t("common.detail.none")),1)],64))]),actions:e(({row:n})=>[l(P,null,{default:e(()=>[l(g,{to:{name:"policy-detail-view",params:{mesh:n.mesh,policyPath:i.path,policy:n.id}}},{default:e(()=>[o(p(t("common.collection.actions.view")),1)]),_:2},1032,["to"])]),_:2},1024)]),_:2},1032,["headers","items","is-selected-row","onResize"])]),_:2},1032,["items","page","page-size","total","onChange"]),a[16]||(a[16]=o()),s.params.policy?(c(),u(T,{key:0},{default:e(({Component:n})=>[l(F,{onClose:h=>s.replace({name:"policy-list-view",params:{mesh:s.params.mesh,policyPath:s.params.policyPath},query:{page:s.params.page,size:s.params.size}})},{default:e(()=>[typeof m<"u"?(c(),u(H(n),{key:0,items:m.items,"policy-type":i},null,8,["items","policy-type"])):f("",!0)]),_:2},1032,["onClose"])]),_:2},1024)):f("",!0)]),_:2},1032,["src"])]),_:2},1024)])]),_:2},1024)]),_:2},1032,["predicate","items"])]),_:1})}}}),ee=q(J,[["__scopeId","data-v-73fadf2e"]]);export{ee as default};
