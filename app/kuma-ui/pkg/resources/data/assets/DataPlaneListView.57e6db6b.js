import{d as _e,I as S,e as p,f as t,a as V,w as I,n as be,J as b,t as _,F as w,h as C,k as A,b as R,cs as te,Z as J,ct as ae,cu as ge,cv as De,r as T,o as r,c as N,y as we,p as se,l as ne,_ as ie,q as ke,cw as Pe,cx as Te,$ as Ce,i as Ee,cy as W,u as Oe,x as Se,X as ee,cz as Ke,cA as Ie,cB as Ae,a1 as Ve,a2 as Ue,cC as Le,a0 as Ne,K,g as Re,cD as Me}from"./index.dbfc69fe.js";import{g as xe}from"./tableDataUtils.76735c08.js";import{D as ze}from"./DataOverview.e98f79d2.js";import{E as Be}from"./EntityTag.76f5c7c9.js";import{Y as He}from"./YamlView.aa325bb9.js";import{_ as $e}from"./EmptyBlock.vue_vue_type_script_setup_true_lang.5dc5cc90.js";import"./ErrorBlock.f52010c3.js";import"./LoadingBlock.vue_vue_type_script_setup_true_lang.f2781028.js";import"./index.58caa11d.js";import"./CodeBlock.83ca39be.js";import"./_commonjsHelpers.712cc82f.js";const oe=[{key:"status",label:"Status"},{key:"name",label:"Name"},{key:"mesh",label:"Mesh"},{key:"type",label:"Type"},{key:"service",label:"Service"},{key:"protocol",label:"Protocol"},{key:"zone",label:"Zone"},{key:"lastConnected",label:"Last Connected"},{key:"lastUpdated",label:"Last Updated"},{key:"totalUpdates",label:"Total Updates"},{key:"dpVersion",label:"Kuma DP version"},{key:"envoyVersion",label:"Envoy version"},{key:"details",label:"Details",hideLabel:!0}],je=["name","details"],qe=oe.filter(e=>!je.includes(e.key)).map(e=>({tableHeaderKey:e.key,label:e.label,isChecked:!1})),le=["status","name","mesh","type","service","protocol","zone","lastUpdated","dpVersion","details"];function Fe(e,a=le){return oe.filter(i=>a.includes(i.key)?e?!0:i.key!=="zone":!1)}function F(e,a){const i=window.history.state;if(i===null)return;const u=i.current.indexOf("?"),n=u>-1?i.current.substring(u):"",c=new URLSearchParams(n);a!=null?c.set(e,String(a)):c.has(e)&&c.delete(e);const g=c.toString(),D=g===""?"":"?"+g;let f="";if(u>-1?f=i.current.substring(0,u)+D:f=i.current+D,i.current!==f){const l=Object.assign({},i);l.current=f,window.history.replaceState(l,"","#"+f)}}const k=e=>(se("data-v-3c7b444a"),e=e(),ne(),e),Ye={class:"entity-summary entity-section-list"},Ge={class:"entity-title"},Qe=k(()=>t("span",{class:"kutil-sr-only"},"Data plane:",-1)),Ze={class:"definition"},Xe=k(()=>t("span",null,"Mesh:",-1)),Je={key:0},We=k(()=>t("h4",null,"Tags",-1)),et={class:"tag-list"},tt={key:1},at=k(()=>t("h4",null,"Dependencies",-1)),st={class:"mt-2 heading-with-icon"},nt=k(()=>t("h4",null,"Insights",-1)),it={class:"entity-section-list"},ot=["data-testid"],lt=k(()=>t("span",null,"Connect time:",-1)),rt=["data-testid"],dt=k(()=>t("span",null,"Disconnect time:",-1)),ct={class:"definition"},pt=k(()=>t("span",null,"Control plane instance ID:",-1)),ut={key:0},mt=k(()=>t("summary",null," Responses (acknowledged / sent) ",-1)),ht=["data-testid"],yt=_e({__name:"DataPlaneEntitySummary",props:{dataPlaneOverview:{type:Object,required:!0}},setup(e){const a=e,i={"Partially degraded":"partially_degraded",Offline:"offline",Online:"online"},u=S(()=>{const{name:l,mesh:d,dataplane:m}=a.dataPlaneOverview;return{type:"Dataplane",name:l,mesh:d,networking:m.networking}}),n=S(()=>te(a.dataPlaneOverview.dataplane)),c=S(()=>{const l=Array.from(a.dataPlaneOverview.dataplaneInsight.subscriptions);return l.reverse(),l.map(d=>{const m=d.connectTime!==void 0?J(d.connectTime):"\u2014",o=d.disconnectTime!==void 0?J(d.disconnectTime):"\u2014",s=Object.entries(d.status).filter(([h])=>!["total","lastUpdateTime"].includes(h)).map(([h,v])=>{var M,x,U,L,E;const P=`${(M=v.responsesAcknowledged)!=null?M:0} / ${(x=v.responsesSent)!=null?x:0}`;return{type:h.toUpperCase(),ratio:P,responsesSent:(U=v.responsesSent)!=null?U:0,responsesAcknowledged:(L=v.responsesAcknowledged)!=null?L:0,responsesRejected:(E=v.responsesRejected)!=null?E:0}});return{subscription:d,formattedConnectDate:m,formattedDisconnectDate:o,statuses:s}})}),g=S(()=>{const{status:l}=ae(a.dataPlaneOverview.dataplane,a.dataPlaneOverview.dataplaneInsight);return ge[i[l]]}),D=S(()=>{const l=De(a.dataPlaneOverview.dataplaneInsight);return l!==null?Object.entries(l).map(([d,m])=>({name:d,version:m})):[]}),f=S(()=>{const{subscriptions:l}=a.dataPlaneOverview.dataplaneInsight;if(l.length===0)return[];const d=[],m=l[l.length-1],o=m.version.envoy,s=m.version.kumaDp,h=o.kumaDpCompatible!==void 0?o.kumaDpCompatible:!0,v=s.kumaCpCompatible!==void 0?s.kumaCpCompatible:!0;if(!h){const P=`Envoy ${o.version} is not supported by Kuma DP ${s.version}.`;d.push(P)}if(!v){const P=`Kuma DP ${s.version} is not supported by this Kuma control plane.`;d.push(P)}return d});return(l,d)=>{const m=T("router-link");return r(),p("div",Ye,[t("section",null,[t("h3",Ge,[Qe,V(m,{to:{name:"data-plane-detail-view",params:{mesh:e.dataPlaneOverview.mesh,dataPlane:e.dataPlaneOverview.name}}},{default:I(()=>[R(_(e.dataPlaneOverview.name),1)]),_:1},8,["to"]),t("div",{class:be(`status status--${b(g).appearance}`),"data-testid":"data-plane-status-badge"},_(b(g).title.toLowerCase()),3)]),t("div",Ze,[Xe,t("span",null,_(e.dataPlaneOverview.mesh),1)])]),b(n).length>0?(r(),p("section",Je,[We,t("div",et,[(r(!0),p(w,null,C(b(n),(o,s)=>(r(),N(Be,{key:s,tag:o},null,8,["tag"]))),128))])])):A("",!0),b(D).length>0?(r(),p("section",tt,[at,(r(!0),p(w,null,C(b(D),(o,s)=>(r(),p("div",{key:s,class:"definition"},[t("span",null,_(o.name)+":",1),t("span",null,_(o.version),1)]))),128)),b(f).length>0?(r(),p(w,{key:0},[t("h5",st,[R(" Warnings "),V(b(we),{class:"ml-1",icon:"warning",color:"var(--black-75)","secondary-color":"var(--yellow-300)",size:"20"})]),(r(!0),p(w,null,C(b(f),(o,s)=>(r(),p("p",{key:s},_(o),1))),128))],64)):A("",!0)])):A("",!0),b(c).length>0?(r(),p(w,{key:2},[t("section",null,[nt,t("div",it,[(r(!0),p(w,null,C(b(c),(o,s)=>(r(),p("div",{key:s},[t("div",{class:"definition","data-testid":`data-plane-connect-time-${s}`},[lt,t("span",null,_(o.formattedConnectDate),1)],8,ot),t("div",{class:"definition","data-testid":`data-plane-disconnect-time-${s}`},[dt,t("span",null,_(o.formattedDisconnectDate),1)],8,rt),t("div",ct,[pt,t("span",null,_(o.subscription.controlPlaneInstanceId),1)]),o.statuses.length>0?(r(),p("details",ut,[mt,(r(!0),p(w,null,C(o.statuses,(h,v)=>(r(),p("div",{key:`${s}-${v}`,class:"definition","data-testid":`data-plane-subscription-status-${s}-${v}`},[t("span",null,_(h.type)+":",1),t("span",null,_(h.ratio),1)],8,ht))),128))])):A("",!0)]))),128))])]),t("section",null,[V(He,{content:b(u),"code-max-height":"250px"},null,8,["content"])])],64)):A("",!0)])}}});const vt=ie(yt,[["__scopeId","data-v-3c7b444a"]]);const ft={name:"DataPlaneListView",dataPlaneTypes:["All","Standard","Gateway (builtin)","Gateway (provided)"],emptyStateMsg:"There are no data plane proxies present.",nsBackButtonRoute:{name:"data-plane-list-view"},dataplaneApiParams:{},components:{DataOverview:ze,DataPlaneEntitySummary:vt,KButton:ke,KDropdownItem:Pe,KDropdownMenu:Te,EmptyBlock:$e},props:{name:{type:String,required:!1,default:null},offset:{type:Number,required:!1,default:0}},data(){return{visibleTableHeaderKeys:le,productName:Ce,isLoading:!0,isEmpty:!1,hasError:!1,tableDataIsEmpty:!1,tableData:{headers:[],data:[]},pageSize:50,next:null,shownTLSTab:!1,rawData:null,filteredDataPlaneType:"All",pageOffset:this.offset,dataPlaneOverview:null}},computed:{...Ee({environment:"config/getEnvironment",queryNamespace:"getItemQueryNamespace",multicluster:"config/getMulticlusterStatus"}),dataplaneWizardRoute(){return this.environment==="universal"?{name:"universal-dataplane"}:{name:"kubernetes-dataplane"}},filteredTableData(){const e=this.tableData.data.filter(i=>this.filteredDataPlaneType==="All"?!0:i.type.toLowerCase()===this.filteredDataPlaneType.toLowerCase()),a=Fe(this.multicluster,this.visibleTableHeaderKeys);return{data:e,headers:a}},columnsDropdownItems(){return qe.filter(e=>this.multicluster?!0:e.tableHeaderKey!=="zone").map(e=>{const a=this.visibleTableHeaderKeys.includes(e.tableHeaderKey);return{...e,isChecked:a}})}},watch:{"$route.params.mesh":function(){this.$route.name==="data-plane-list-view"&&(this.isLoading=!0,this.isEmpty=!1,this.hasError=!1,this.tableDataIsEmpty=!1,this.loadData(0))}},created(){const e=W.get("dpVisibleTableHeaderKeys");Array.isArray(e)&&(this.visibleTableHeaderKeys=e)},beforeMount(){this.loadData(this.offset)},methods:{stopPropagatingClickEvent(e){e.stopPropagation()},updateVisibleTableHeaders(e,a){const i=e.target,u=this.visibleTableHeaderKeys.findIndex(n=>n===a);i.checked&&u===-1?this.visibleTableHeaderKeys.push(a):!i.checked&&u>-1&&this.visibleTableHeaderKeys.splice(u,1),W.set("dpVisibleTableHeaderKeys",Array.from(new Set(this.visibleTableHeaderKeys)))},onCreateClick(){Oe.logger.info(Se.CREATE_DATA_PLANE_PROXY_CLICKED)},getEmptyState(){return{title:"No Data",message:this.$options.emptyStateMsg}},parseData(e){var G,Q,Z;const{dataplane:a={},dataplaneInsight:i={}}=e,{name:u="",mesh:n=""}=e,{subscriptions:c=[]}=i,g={name:"data-plane-detail-view",params:{mesh:n,dataPlane:u}},D={name:"mesh-child",params:{mesh:n}},f=["kuma.io/protocol","kuma.io/service","kuma.io/zone"],l=te(a).filter(y=>f.includes(y.label)),d=(G=l.find(y=>y.label==="kuma.io/service"))==null?void 0:G.value,m=(Q=l.find(y=>y.label==="kuma.io/protocol"))==null?void 0:Q.value,o=(Z=l.find(y=>y.label==="kuma.io/zone"))==null?void 0:Z.value;let s;d!==void 0&&(s={name:"service-insight-detail-view",params:{mesh:n,service:d}});const{status:h}=ae(a,i),{totalUpdates:v,totalRejectedUpdates:P,dpVersion:M,envoyVersion:x,selectedTime:U,selectedUpdateTime:L,version:E}=c.reduce((y,$)=>{const{status:re={},connectTime:de,version:X={}}=$,{total:ce={},lastUpdateTime:pe}=re,{responsesSent:ue="0",responsesRejected:me="0"}=ce,{kumaDp:he={},envoy:ye={}}=X,{version:ve}=he,{version:fe}=ye;let{selectedTime:z,selectedUpdateTime:B}=y;const j=Date.parse(de),q=Date.parse(pe);return j&&(!z||j>z)&&(z=j),q&&(!B||q>B)&&(B=q),{totalUpdates:y.totalUpdates+parseInt(ue,10),totalRejectedUpdates:y.totalRejectedUpdates+parseInt(me,10),dpVersion:ve||y.dpVersion,envoyVersion:fe||y.envoyVersion,selectedTime:z,selectedUpdateTime:B,version:X||y.version}},{totalUpdates:0,totalRejectedUpdates:0,dpVersion:"\u2014",envoyVersion:"\u2014",selectedTime:NaN,selectedUpdateTime:NaN,version:{}}),O={name:u,nameRoute:g,mesh:n,meshRoute:D,zone:o!=null?o:"\u2014",service:d!=null?d:"\u2014",serviceInsightRoute:s,protocol:m!=null?m:"\u2014",status:h,totalUpdates:v,totalRejectedUpdates:P,dpVersion:M,envoyVersion:x,warnings:[],unsupportedEnvoyVersion:!1,unsupportedKumaDPVersion:!1,kumaDpAndKumaCpMismatch:!1,lastUpdated:L?ee(new Date(L).toUTCString()):"\u2014",lastConnected:U?ee(new Date(U).toUTCString()):"\u2014",type:Ke(a)},{kind:H}=Ie(E);switch(H!==Ae&&O.warnings.push(H),H){case Ue:O.unsupportedEnvoyVersion=!0;break;case Ve:O.unsupportedKumaDPVersion=!0;break}return this.multicluster&&l.find($=>$.label===Le)&&typeof E.kumaDp.kumaCpCompatible=="boolean"&&!E.kumaDp.kumaCpCompatible&&(O.warnings.push(Ne),O.kumaDpAndKumaCpMismatch=!0),O},async loadData(e){var u;this.isLoading=!0,this.pageOffset=e,F("offset",e>0?e:null);const a=this.$route.params.mesh||null,i=this.$route.query.ns||null;try{const{data:n,next:c}=await xe({getSingleEntity:K.getDataplaneOverviewFromMesh.bind(K),getAllEntities:K.getAllDataplaneOverviews.bind(K),getAllEntitiesFromMesh:K.getAllDataplaneOverviewsFromMesh.bind(K),size:this.pageSize,offset:e,mesh:a,query:i,params:{...this.$options.dataplaneApiParams}});if(n.length>0){this.next=c,this.rawData=n,this.selectDataPlaneOverview((u=this.name)!=null?u:n[0].name);const g=await Promise.all(n.map(D=>this.parseData(D)));this.tableData.data=g,this.tableDataIsEmpty=!1,this.isEmpty=!1}else this.selectDataPlaneOverview(null),this.tableData.data=[],this.tableDataIsEmpty=!0,this.isEmpty=!0}catch(n){this.hasError=!0,this.isEmpty=!0,console.error(n)}finally{this.isLoading=!1}},async selectDataPlaneOverview(e){var a;e?(this.dataPlaneOverview=(a=this.rawData.find(i=>i.name===e))!=null?a:this.rawData[0],F("name",this.dataPlaneOverview.name)):(this.dataPlaneOverview=null,F("name",null))}}},Y=e=>(se("data-v-0b3e029f"),e=e(),ne(),e),_t={class:"data-planes-container"},bt={class:"data-planes-content component-frame"},gt=Y(()=>t("label",{for:"data-planes-type-filter",class:"mr-2"}," Type: ",-1)),Dt=["value"],wt=["for"],kt=["id","checked","onChange"],Pt=Y(()=>t("span",{class:"custom-control-icon"}," + ",-1)),Tt=Y(()=>t("span",{class:"custom-control-icon"}," \u2190 ",-1)),Ct={class:"data-planes-sidebar component-frame"};function Et(e,a,i,u,n,c){var o;const g=T("KDropdownItem"),D=T("KDropdownMenu"),f=T("KButton"),l=T("DataOverview"),d=T("DataPlaneEntitySummary"),m=T("EmptyBlock");return r(),p("div",_t,[t("div",bt,[V(l,{"selected-entity-name":(o=n.dataPlaneOverview)==null?void 0:o.name,"page-size":n.pageSize,"has-error":n.hasError,"is-loading":n.isLoading,"empty-state":c.getEmptyState(),"table-data":c.filteredTableData,"table-data-is-empty":n.tableDataIsEmpty,"show-details":"",next:n.next,"page-offset":n.pageOffset,onTableAction:a[2]||(a[2]=s=>c.selectDataPlaneOverview(s.name)),onLoadData:a[3]||(a[3]=s=>c.loadData(s))},{additionalControls:I(()=>[t("div",null,[gt,Re(t("select",{id:"data-planes-type-filter","onUpdate:modelValue":a[0]||(a[0]=s=>n.filteredDataPlaneType=s),"data-testid":"data-planes-type-filter"},[(r(!0),p(w,null,C(e.$options.dataPlaneTypes,(s,h)=>(r(),p("option",{key:h,value:s},_(s),9,Dt))),128))],512),[[Me,n.filteredDataPlaneType]])]),V(D,{label:"Columns",icon:"cogwheel","button-appearance":"outline"},{items:I(()=>[t("div",{onClick:a[1]||(a[1]=(...s)=>c.stopPropagatingClickEvent&&c.stopPropagatingClickEvent(...s))},[(r(!0),p(w,null,C(c.columnsDropdownItems,(s,h)=>(r(),N(g,{key:h,class:"table-header-selector-item",item:s},{default:I(()=>[t("label",{for:`data-plane-table-header-checkbox-${h}`,class:"k-checkbox table-header-selector-item-checkbox"},[t("input",{id:`data-plane-table-header-checkbox-${h}`,checked:s.isChecked,type:"checkbox",class:"k-input",onChange:v=>c.updateVisibleTableHeaders(v,s.tableHeaderKey)},null,40,kt),R(" "+_(s.label),1)],8,wt)]),_:2},1032,["item"]))),128))])]),_:1}),V(f,{class:"add-dp-button",appearance:"primary",to:c.dataplaneWizardRoute,"data-testid":"data-plane-create-data-plane-button",onClick:c.onCreateClick},{default:I(()=>[Pt,R(" Create data plane proxy ")]),_:1},8,["to","onClick"]),e.$route.query.ns?(r(),N(f,{key:0,appearance:"primary",to:e.$options.nsBackButtonRoute,"data-testid":"data-plane-ns-back-button"},{default:I(()=>[Tt,R(" View All ")]),_:1},8,["to"])):A("",!0)]),_:1},8,["selected-entity-name","page-size","has-error","is-loading","empty-state","table-data","table-data-is-empty","next","page-offset"])]),t("div",Ct,[n.dataPlaneOverview!==null?(r(),N(d,{key:0,"data-plane-overview":n.dataPlaneOverview},null,8,["data-plane-overview"])):(r(),N(m,{key:1}))])])}const xt=ie(ft,[["render",Et],["__scopeId","data-v-0b3e029f"]]);export{xt as default};
