import{d as P,r as m,o as i,m as y,w as e,b as n,s as u,t as d,e as t,c as f,F as h,v as _,n as q,T as A,U as c,S as I,p as S,ag as Q,q as k,Y as U,_ as O}from"./index-ChH5weWG.js";import{T as K}from"./TagList-BJLwm2eA.js";import{_ as z}from"./ResourceCodeBlock.vue_vue_type_script_setup_true_lang-rvKSlgEf.js";const Y={class:"stack-with-borders"},Z={class:"stack-with-borders"},$=P({__name:"DataPlaneSummaryView",props:{items:{},routeName:{}},setup(T){const C=T;return(j,a)=>{const E=m("XEmptyState"),X=m("RouteTitle"),w=m("XAction"),V=m("XSelect"),g=m("XLayout"),D=m("XIcon"),b=m("DataCollection"),F=m("XCopyButton"),M=m("DataLoader"),B=m("AppView"),v=m("RouteView");return i(),y(v,{name:C.routeName,params:{mesh:"",proxy:"",codeSearch:"",codeFilter:!1,codeRegExp:!1,format:String}},{default:e(({route:s,t:l,uri:L,can:N})=>[n(b,{items:C.items,predicate:x=>x.id===s.params.proxy},{empty:e(()=>[n(E,null,{title:e(()=>[u("h2",null,d(l("common.collection.summary.empty_title",{type:"Data Plane Proxy"})),1)]),default:e(()=>[a[0]||(a[0]=t()),u("p",null,d(l("common.collection.summary.empty_message",{type:"Data Plane Proxy"})),1)]),_:2},1024)]),default:e(({items:x})=>[(i(!0),f(h,null,_([x[0]],r=>(i(),y(B,{key:r.id},{title:e(()=>[u("h2",{class:q(`type-${r.dataplaneType}`)},[n(w,{to:{name:"data-plane-detail-view",params:{proxy:r.id}}},{default:e(()=>[n(X,{title:l("data-planes.routes.item.title",{name:r.name})},null,8,["title"])]),_:2},1032,["to"])],2)]),default:e(()=>[a[19]||(a[19]=t()),n(g,null,{default:e(()=>[u("header",null,[n(g,{type:"separated",size:"max"},{default:e(()=>[u("h3",null,d(l("data-planes.routes.item.config")),1),a[1]||(a[1]=t()),(i(),f(h,null,_([["structured","universal","k8s"]],p=>u("div",{key:typeof p},[n(V,{label:l("data-planes.routes.item.format"),selected:s.params.format,onChange:o=>{s.update({format:o})},onVnodeBeforeMount:o=>{var R;return((R=o==null?void 0:o.props)==null?void 0:R.selected)&&p.includes(o.props.selected)&&o.props.selected!==s.params.format&&s.update({format:o.props.selected})}},A({_:2},[_(p,o=>({name:`${o}-option`,fn:e(()=>[t(d(l(`data-planes.routes.item.formats.${o}`)),1)])}))]),1032,["label","selected","onChange","onVnodeBeforeMount"])])),64))]),_:2},1024)])]),_:2},1024),a[20]||(a[20]=t()),s.params.format==="structured"?(i(),y(g,{key:0,type:"stack","data-testid":"structured-view"},{default:e(()=>[u("div",Y,[n(c,{layout:"horizontal"},{title:e(()=>[t(d(l("http.api.property.status")),1)]),body:e(()=>[n(g,{type:"separated"},{default:e(()=>[n(I,{status:r.status},null,8,["status"]),a[2]||(a[2]=t()),r.dataplaneType==="standard"?(i(),y(b,{key:0,items:r.dataplane.networking.inbounds,predicate:p=>p.state!=="Ready",empty:!1},{default:e(({items:p})=>[n(D,{name:"info",color:S(Q)},{default:e(()=>[u("ul",null,[(i(!0),f(h,null,_(p,o=>(i(),f("li",{key:`${o.service}:${o.port}`},d(l("data-planes.routes.item.unhealthy_inbound",{service:o.service,port:o.port})),1))),128))])]),_:2},1032,["color"])]),_:2},1032,["items","predicate"])):k("",!0)]),_:2},1024)]),_:2},1024),a[10]||(a[10]=t()),n(c,{layout:"horizontal"},{title:e(()=>a[4]||(a[4]=[t(`
                      Type
                    `)])),body:e(()=>[t(d(l(`data-planes.type.${r.dataplaneType}`)),1)]),_:2},1024),a[11]||(a[11]=t()),r.namespace.length>0?(i(),y(c,{key:0,layout:"horizontal"},{title:e(()=>[t(d(l("data-planes.routes.item.namespace")),1)]),body:e(()=>[t(d(r.namespace),1)]),_:2},1024)):k("",!0),a[12]||(a[12]=t()),N("use zones")&&r.zone?(i(),y(c,{key:1,layout:"horizontal"},{title:e(()=>a[7]||(a[7]=[t(`
                      Zone
                    `)])),body:e(()=>[n(w,{to:{name:"zone-cp-detail-view",params:{zone:r.zone}}},{default:e(()=>[t(d(r.zone),1)]),_:2},1032,["to"])]),_:2},1024)):k("",!0),a[13]||(a[13]=t()),n(c,{layout:"horizontal"},{title:e(()=>[t(d(l("http.api.property.modificationTime")),1)]),body:e(()=>[t(d(l("common.formats.datetime",{value:Date.parse(r.modificationTime)})),1)]),_:2},1024)]),a[18]||(a[18]=t()),r.dataplane.networking.gateway?(i(),y(g,{key:0,type:"stack"},{default:e(()=>[u("h3",null,d(l("data-planes.routes.item.gateway")),1),a[17]||(a[17]=t()),u("div",Z,[n(c,{layout:"horizontal"},{title:e(()=>[t(d(l("http.api.property.tags")),1)]),body:e(()=>[n(K,{alignment:"right",tags:r.dataplane.networking.gateway.tags},null,8,["tags"])]),_:2},1024),a[16]||(a[16]=t()),n(c,{layout:"horizontal"},{title:e(()=>[t(d(l("http.api.property.address")),1)]),body:e(()=>[n(F,{text:`${r.dataplane.networking.address}`},null,8,["text"])]),_:2},1024)])]),_:2},1024)):k("",!0)]),_:2},1024)):s.params.format==="universal"?(i(),y(z,{key:1,"data-testid":"codeblock-yaml-universal",language:"yaml",resource:r.config,"show-k8s-copy-button":!1,"is-searchable":"",query:s.params.codeSearch,"is-filter-mode":s.params.codeFilter,"is-reg-exp-mode":s.params.codeRegExp,onQueryChange:p=>s.update({codeSearch:p}),onFilterModeChange:p=>s.update({codeFilter:p}),onRegExpModeChange:p=>s.update({codeRegExp:p})},null,8,["resource","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])):(i(),y(M,{key:2,src:L(S(U),"/meshes/:mesh/dataplanes/:name/as/kubernetes",{mesh:s.params.mesh,name:s.params.proxy})},{default:e(({data:p})=>[n(z,{"data-testid":"codeblock-yaml-k8s",language:"yaml",resource:p,"show-k8s-copy-button":!1,"is-searchable":"",query:s.params.codeSearch,"is-filter-mode":s.params.codeFilter,"is-reg-exp-mode":s.params.codeRegExp,onQueryChange:o=>s.update({codeSearch:o}),onFilterModeChange:o=>s.update({codeFilter:o}),onRegExpModeChange:o=>s.update({codeRegExp:o})},null,8,["resource","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])]),_:2},1032,["src"]))]),_:2},1024))),128))]),_:2},1032,["items","predicate"])]),_:1},8,["name"])}}}),W=O($,[["__scopeId","data-v-c0d1d54d"]]);export{W as default};
