import{d as A,r as i,o as d,q as u,w as e,b as o,m,t as r,e as a,c as C,K as T,L as w,n as M,Q as q,R as y,S as I,p as b,ae as Q,s as g,V as K}from"./index-CKQWVGYP.js";import{T as O}from"./TagList-C2lT_Az9.js";import{_ as U}from"./ResourceCodeBlock.vue_vue_type_script_setup_true_lang-BDFzF6sj.js";const Z={class:"stack-with-borders"},j={class:"stack-with-borders"},Y=A({__name:"DataPlaneSummaryView",props:{items:{},routeName:{}},setup(R){const h=R;return(G,t)=>{const X=i("XEmptyState"),E=i("RouteTitle"),x=i("XAction"),v=i("XSelect"),c=i("XLayout"),D=i("XIcon"),z=i("DataCollection"),V=i("XCopyButton"),P=i("DataSource"),$=i("AppView"),B=i("RouteView");return d(),u(B,{name:h.routeName,params:{mesh:"",dataPlane:"",codeSearch:"",codeFilter:!1,codeRegExp:!1,format:"structured"}},{default:e(({route:p,t:l,uri:L,can:N})=>[o(z,{items:h.items,predicate:_=>_.id===p.params.dataPlane},{empty:e(()=>[o(X,null,{title:e(()=>[m("h2",null,r(l("common.collection.summary.empty_title",{type:"Data Plane Proxy"})),1)]),default:e(()=>[t[0]||(t[0]=a()),m("p",null,r(l("common.collection.summary.empty_message",{type:"Data Plane Proxy"})),1)]),_:2},1024)]),default:e(({items:_})=>[(d(!0),C(T,null,w([_[0]],n=>(d(),u($,{key:n.id},{title:e(()=>[m("h2",{class:M(`type-${n.dataplaneType}`)},[o(x,{to:{name:"data-plane-detail-view",params:{dataPlane:n.id}}},{default:e(()=>[o(E,{title:l("data-planes.routes.item.title",{name:n.name})},null,8,["title"])]),_:2},1032,["to"])],2)]),default:e(()=>[t[19]||(t[19]=a()),o(c,null,{default:e(()=>[m("header",null,[o(c,{type:"separated",size:"max"},{default:e(()=>[m("h3",null,r(l("data-planes.routes.item.config")),1),t[1]||(t[1]=a()),m("div",null,[o(v,{label:l("data-planes.routes.item.format"),selected:p.params.format,onChange:s=>{p.update({format:s})}},q({_:2},[w(["structured","yaml"],s=>({name:`${s}-option`,fn:e(()=>[a(r(l(`data-planes.routes.item.formats.${s}`)),1)])}))]),1032,["label","selected","onChange"])])]),_:2},1024)])]),_:2},1024),t[20]||(t[20]=a()),p.params.format==="structured"?(d(),u(c,{key:0,type:"stack","data-testid":"structured-view"},{default:e(()=>[m("div",Z,[o(y,{layout:"horizontal"},{title:e(()=>[a(r(l("http.api.property.status")),1)]),body:e(()=>[o(c,{type:"separated"},{default:e(()=>[o(I,{status:n.status},null,8,["status"]),t[2]||(t[2]=a()),n.dataplaneType==="standard"?(d(),u(z,{key:0,items:n.dataplane.networking.inbounds,predicate:s=>s.state!=="Ready",empty:!1},{default:e(({items:s})=>[o(D,{name:"info",color:b(Q)},{default:e(()=>[m("ul",null,[(d(!0),C(T,null,w(s,f=>(d(),C("li",{key:`${f.service}:${f.port}`},r(l("data-planes.routes.item.unhealthy_inbound",{service:f.service,port:f.port})),1))),128))])]),_:2},1032,["color"])]),_:2},1032,["items","predicate"])):g("",!0)]),_:2},1024)]),_:2},1024),t[10]||(t[10]=a()),o(y,{layout:"horizontal"},{title:e(()=>t[4]||(t[4]=[a(`
                      Type
                    `)])),body:e(()=>[a(r(l(`data-planes.type.${n.dataplaneType}`)),1)]),_:2},1024),t[11]||(t[11]=a()),n.namespace.length>0?(d(),u(y,{key:0,layout:"horizontal"},{title:e(()=>[a(r(l("data-planes.routes.item.namespace")),1)]),body:e(()=>[a(r(n.namespace),1)]),_:2},1024)):g("",!0),t[12]||(t[12]=a()),N("use zones")&&n.zone?(d(),u(y,{key:1,layout:"horizontal"},{title:e(()=>t[7]||(t[7]=[a(`
                      Zone
                    `)])),body:e(()=>[o(x,{to:{name:"zone-cp-detail-view",params:{zone:n.zone}}},{default:e(()=>[a(r(n.zone),1)]),_:2},1032,["to"])]),_:2},1024)):g("",!0),t[13]||(t[13]=a()),o(y,{layout:"horizontal"},{title:e(()=>[a(r(l("http.api.property.modificationTime")),1)]),body:e(()=>[a(r(l("common.formats.datetime",{value:Date.parse(n.modificationTime)})),1)]),_:2},1024)]),t[18]||(t[18]=a()),n.dataplane.networking.gateway?(d(),u(c,{key:0,type:"stack"},{default:e(()=>[m("h3",null,r(l("data-planes.routes.item.gateway")),1),t[17]||(t[17]=a()),m("div",j,[o(y,{layout:"horizontal"},{title:e(()=>[a(r(l("http.api.property.tags")),1)]),body:e(()=>[o(O,{alignment:"right",tags:n.dataplane.networking.gateway.tags},null,8,["tags"])]),_:2},1024),t[16]||(t[16]=a()),o(y,{layout:"horizontal"},{title:e(()=>[a(r(l("http.api.property.address")),1)]),body:e(()=>[o(V,{text:`${n.dataplane.networking.address}`},null,8,["text"])]),_:2},1024)])]),_:2},1024)):g("",!0)]),_:2},1024)):(d(),u(c,{key:1,type:"stack"},{default:e(()=>[o(U,{resource:n.config,language:"yaml","is-searchable":"",query:p.params.codeSearch,"is-filter-mode":p.params.codeFilter,"is-reg-exp-mode":p.params.codeRegExp,onQueryChange:s=>p.update({codeSearch:s}),onFilterModeChange:s=>p.update({codeFilter:s}),onRegExpModeChange:s=>p.update({codeRegExp:s})},{default:e(({copy:s,copying:f})=>[f?(d(),u(P,{key:0,src:L(b(K),"/meshes/:mesh/dataplanes/:name/as/kubernetes",{mesh:p.params.mesh,name:p.params.dataPlane},{cacheControl:"no-store"}),onChange:k=>{s(S=>S(k))},onError:k=>{s((S,F)=>F(k))}},null,8,["src","onChange","onError"])):g("",!0)]),_:2},1032,["resource","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])]),_:2},1024))]),_:2},1024))),128))]),_:2},1032,["items","predicate"])]),_:1},8,["name"])}}});export{Y as default};
