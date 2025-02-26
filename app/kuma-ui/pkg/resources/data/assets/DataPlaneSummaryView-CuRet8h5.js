import{d as A,r as i,o as d,m as u,w as e,b as o,s as m,t as l,e as a,c as C,F as T,v as x,n as I,T as M,U as y,S as q,p as b,ah as U,q as _,Y as O,_ as Q}from"./index-C-Llvxgw.js";import{T as K}from"./TagList-CVIOP13G.js";import{_ as Y}from"./ResourceCodeBlock.vue_vue_type_script_setup_true_lang-D7DuEGVs.js";const Z={class:"stack-with-borders"},j={class:"stack-with-borders"},G=A({__name:"DataPlaneSummaryView",props:{items:{},routeName:{}},setup(v){const w=v;return(H,t)=>{const X=i("XEmptyState"),E=i("RouteTitle"),h=i("XAction"),R=i("XSelect"),c=i("XLayout"),D=i("XIcon"),S=i("DataCollection"),V=i("XCopyButton"),$=i("DataSource"),B=i("AppView"),F=i("RouteView");return d(),u(F,{name:w.routeName,params:{mesh:"",proxy:"",codeSearch:"",codeFilter:!1,codeRegExp:!1,format:String}},{default:e(({route:p,t:r,uri:N,can:L})=>[o(S,{items:w.items,predicate:g=>g.id===p.params.proxy},{empty:e(()=>[o(X,null,{title:e(()=>[m("h2",null,l(r("common.collection.summary.empty_title",{type:"Data Plane Proxy"})),1)]),default:e(()=>[t[0]||(t[0]=a()),m("p",null,l(r("common.collection.summary.empty_message",{type:"Data Plane Proxy"})),1)]),_:2},1024)]),default:e(({items:g})=>[(d(!0),C(T,null,x([g[0]],n=>(d(),u(B,{key:n.id},{title:e(()=>[m("h2",{class:I(`type-${n.dataplaneType}`)},[o(h,{to:{name:"data-plane-detail-view",params:{proxy:n.id}}},{default:e(()=>[o(E,{title:r("data-planes.routes.item.title",{name:n.name})},null,8,["title"])]),_:2},1032,["to"])],2)]),default:e(()=>[t[19]||(t[19]=a()),o(c,null,{default:e(()=>[m("header",null,[o(c,{type:"separated",size:"max"},{default:e(()=>[m("h3",null,l(r("data-planes.routes.item.config")),1),t[1]||(t[1]=a()),m("div",null,[o(R,{label:r("data-planes.routes.item.format"),selected:p.params.format,onChange:s=>{p.update({format:s})}},M({_:2},[x(["structured","yaml"],s=>({name:`${s}-option`,fn:e(()=>[a(l(r(`data-planes.routes.item.formats.${s}`)),1)])}))]),1032,["label","selected","onChange"])])]),_:2},1024)])]),_:2},1024),t[20]||(t[20]=a()),p.params.format==="structured"?(d(),u(c,{key:0,type:"stack","data-testid":"structured-view"},{default:e(()=>[m("div",Z,[o(y,{layout:"horizontal"},{title:e(()=>[a(l(r("http.api.property.status")),1)]),body:e(()=>[o(c,{type:"separated"},{default:e(()=>[o(q,{status:n.status},null,8,["status"]),t[2]||(t[2]=a()),n.dataplaneType==="standard"?(d(),u(S,{key:0,items:n.dataplane.networking.inbounds,predicate:s=>s.state!=="Ready",empty:!1},{default:e(({items:s})=>[o(D,{name:"info",color:b(U)},{default:e(()=>[m("ul",null,[(d(!0),C(T,null,x(s,f=>(d(),C("li",{key:`${f.service}:${f.port}`},l(r("data-planes.routes.item.unhealthy_inbound",{service:f.service,port:f.port})),1))),128))])]),_:2},1032,["color"])]),_:2},1032,["items","predicate"])):_("",!0)]),_:2},1024)]),_:2},1024),t[10]||(t[10]=a()),o(y,{layout:"horizontal"},{title:e(()=>t[4]||(t[4]=[a(`
                      Type
                    `)])),body:e(()=>[a(l(r(`data-planes.type.${n.dataplaneType}`)),1)]),_:2},1024),t[11]||(t[11]=a()),n.namespace.length>0?(d(),u(y,{key:0,layout:"horizontal"},{title:e(()=>[a(l(r("data-planes.routes.item.namespace")),1)]),body:e(()=>[a(l(n.namespace),1)]),_:2},1024)):_("",!0),t[12]||(t[12]=a()),L("use zones")&&n.zone?(d(),u(y,{key:1,layout:"horizontal"},{title:e(()=>t[7]||(t[7]=[a(`
                      Zone
                    `)])),body:e(()=>[o(h,{to:{name:"zone-cp-detail-view",params:{zone:n.zone}}},{default:e(()=>[a(l(n.zone),1)]),_:2},1032,["to"])]),_:2},1024)):_("",!0),t[13]||(t[13]=a()),o(y,{layout:"horizontal"},{title:e(()=>[a(l(r("http.api.property.modificationTime")),1)]),body:e(()=>[a(l(r("common.formats.datetime",{value:Date.parse(n.modificationTime)})),1)]),_:2},1024)]),t[18]||(t[18]=a()),n.dataplane.networking.gateway?(d(),u(c,{key:0,type:"stack"},{default:e(()=>[m("h3",null,l(r("data-planes.routes.item.gateway")),1),t[17]||(t[17]=a()),m("div",j,[o(y,{layout:"horizontal"},{title:e(()=>[a(l(r("http.api.property.tags")),1)]),body:e(()=>[o(K,{alignment:"right",tags:n.dataplane.networking.gateway.tags},null,8,["tags"])]),_:2},1024),t[16]||(t[16]=a()),o(y,{layout:"horizontal"},{title:e(()=>[a(l(r("http.api.property.address")),1)]),body:e(()=>[o(V,{text:`${n.dataplane.networking.address}`},null,8,["text"])]),_:2},1024)])]),_:2},1024)):_("",!0)]),_:2},1024)):(d(),u(c,{key:1,type:"stack"},{default:e(()=>[o(Y,{resource:n.config,language:"yaml","is-searchable":"",query:p.params.codeSearch,"is-filter-mode":p.params.codeFilter,"is-reg-exp-mode":p.params.codeRegExp,onQueryChange:s=>p.update({codeSearch:s}),onFilterModeChange:s=>p.update({codeFilter:s}),onRegExpModeChange:s=>p.update({codeRegExp:s})},{default:e(({copy:s,copying:f})=>[f?(d(),u($,{key:0,src:N(b(O),"/meshes/:mesh/dataplanes/:name/as/kubernetes",{mesh:p.params.mesh,name:p.params.proxy},{cacheControl:"no-store"}),onChange:k=>{s(z=>z(k))},onError:k=>{s((z,P)=>P(k))}},null,8,["src","onChange","onError"])):_("",!0)]),_:2},1032,["resource","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])]),_:2},1024))]),_:2},1024))),128))]),_:2},1032,["items","predicate"])]),_:1},8,["name"])}}}),te=Q(G,[["__scopeId","data-v-44b8f766"]]);export{te as default};
